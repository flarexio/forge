package forge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/filters"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
)

type Service interface {

	// ImageDescription retrieves the description of a Docker image from Docker Hub.
	// i.e. https://hub.docker.com/v2/repositories/<image>
	ImageDescription(ctx context.Context, image string) (description string, err error)

	// ListImages lists all available Docker images on the host.
	// i.e. sudo docker images
	ListImages(ctx context.Context, page int, pageSize int) (images []string, err error)

	// PullImage pulls a Docker image from a registry.
	// i.e. sudo docker pull <image>
	PullImage(ctx context.Context, image string) error

	// ListContainers lists all containers in the current workspace.
	// i.e. sudo docker ps -a --filter "label=workspace=<workspaceID>"
	ListContainers(ctx context.Context) (containers []*Container, err error)

	// RunContainerOnce runs a container to execute a command and then removes the container.
	// i.e. sudo docker run --rm alpine:latest echo hello world
	RunContainerOnce(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (output string, err error)

	// RunContainer runs a container with interactive shell.
	// i.e. sudo docker run -id alpine:latest /bin/sh
	RunContainer(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (containerID string, err error)

	// SendToContainer sends input to the container's stdin.
	SendToContainer(ctx context.Context, containerID string, input string) error

	// LogsContainer retrieves the logs of a container.
	// i.e. sudo docker logs --tail <lines> <containerID>
	LogsContainer(ctx context.Context, containerID string, since time.Time, tail ...int) (output string, err error)

	// ExecCommand executes a command in a running container.
	// i.e. sudo docker exec <containerID> <command>
	ExecCommand(ctx context.Context, containerID string, cmd ...string) (output string, err error)

	// Wait waits for the specified duration.
	Wait(ctx context.Context, timeout time.Duration)

	// SendAndRead sends input to the container's stdin and reads the output.
	SendAndRead(ctx context.Context, containerID string, input string, wait time.Duration) (output string, err error)

	// RemoveContainer removes a container.
	// i.e. sudo docker rm -f <containerID>
	RemoveContainer(ctx context.Context, containerID string) error

	// RemoveAllContainers removes all containers in the current workspace.
	RemoveAllContainers(ctx context.Context) error

	// Close closes the service and cleans up any resources.
	Close(ctx context.Context) error
}

type ServiceMiddleware func(Service) Service

func NewService(path string) (Service, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	networkName := "forge"

	svc := &service{
		path:        path,
		networkName: networkName,
		cli:         cli,
		sessions:    make(map[string]*client.HijackedResponse),
	}

	ctx := context.Background()
	if err := svc.ensureForgeNetwork(ctx, networkName); err != nil {
		return nil, err
	}

	return svc, nil
}

type service struct {
	path        string
	networkName string
	cli         *client.Client
	sessions    map[string]*client.HijackedResponse
	sync.RWMutex
}

func (svc *service) ensureForgeNetwork(ctx context.Context, networkName string) error {
	networks, err := svc.cli.NetworkList(ctx, client.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})

	if err != nil {
		return err
	}

	if len(networks) > 0 {
		return nil
	}

	_, err = svc.cli.NetworkCreate(ctx, networkName, client.NetworkCreateOptions{
		Driver: "bridge",
	})

	return err
}

// workspaceVolume creates or gets the workspace volume
func (svc *service) workspaceVolume(ctx context.Context, workspaceID string) (string, error) {
	volumeName := "workspace-" + workspaceID

	// Check if volume already exists
	_, err := svc.cli.VolumeInspect(ctx, volumeName)
	if err == nil {
		return volumeName, nil
	}

	// Create workspace directory on host
	workspaceDir := filepath.Join(svc.path, "workspaces", workspaceID)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return "", err
	}

	// Create volume with bind mount to host directory
	_, err = svc.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name:   volumeName,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "none",
			"device": workspaceDir,
			"o":      "bind",
		},
		Labels: map[string]string{
			"workspace": workspaceID,
		},
	})

	if err != nil {
		return "", err
	}

	return volumeName, nil
}

func (svc *service) ImageDescription(ctx context.Context, image string) (description string, err error) {
	if colonIndex := strings.Index(image, ":"); colonIndex != -1 {
		image = image[:colonIndex]
	}

	if !strings.Contains(image, "/") {
		image = "library/" + image
	}

	url := "https://hub.docker.com/v2/repositories/" + image

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	var result struct {
		FullDescription string `json:"full_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.FullDescription, nil
}

func (svc *service) ListImages(ctx context.Context, page int, pageSize int) ([]string, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 20
	}

	summaries, err := svc.cli.ImageList(ctx, client.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	var images []string
	for _, img := range summaries {
		if len(img.RepoTags) > 0 {
			images = append(images, img.RepoTags...)
		} else {
			images = append(images, img.ID)
		}
	}

	// Paginate the results
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}

	end := start + pageSize
	if end > len(images) {
		end = len(images)
	}

	images = images[start:end]

	return images, nil
}

func (svc *service) PullImage(ctx context.Context, image string) error {
	reader, err := svc.cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read the output to ensure the image is pulled completely
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (svc *service) ListContainers(ctx context.Context) ([]*Container, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return nil, ErrNoWorkspaceID
	}

	if !isValidWorkspaceID(workspaceID) {
		return nil, ErrWorkspaceNotFound
	}

	filters := filters.NewArgs()
	filters.Add("label", "workspace="+workspaceID)

	summaries, err := svc.cli.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: filters,
	})

	if err != nil {
		return nil, err
	}

	containers := make([]*Container, len(summaries))
	for i, s := range summaries {
		c := &Container{
			ID:      s.ID,
			Image:   s.Image,
			Command: s.Command,
			Status:  s.Status,
			Names:   s.Names,
		}

		containers[i] = c
	}

	return containers, nil
}

func (svc *service) RunContainerOnce(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (string, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return "", ErrNoWorkspaceID
	}

	if !isValidWorkspaceID(workspaceID) {
		return "", ErrWorkspaceNotFound
	}

	_, err := svc.cli.ImageInspect(ctx, image)
	if err != nil {
		_, err := svc.cli.ImagePull(ctx, image, client.ImagePullOptions{})
		if err != nil {
			return "", err
		}
	}

	binds := []string{
		"/etc/localtime:/etc/localtime:ro",
	}

	if mountPath != "" {
		volumeName, err := svc.workspaceVolume(ctx, workspaceID)
		if err != nil {
			return "", err
		}

		binds = append(binds, volumeName+":"+mountPath)
	}

	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	networkName := svc.networkName
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}

	config := &container.Config{
		Image: image,
		Cmd:   cmd,
		Labels: map[string]string{
			"workspace": workspaceID,
		},
	}

	if workDir != "" {
		config.WorkingDir = workDir
	}

	execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := svc.cli.ContainerCreate(execCtx, config, hostConfig, networkConfig, nil, "")
	if err != nil {
		return "", err
	}

	defer func() {
		svc.cli.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true})
	}()

	containerID := resp.ID

	if err := svc.cli.ContainerStart(execCtx, containerID,
		client.ContainerStartOptions{},
	); err != nil {
		return "", err
	}

	// 使用 channel 來處理容器等待和讀取日誌的逾時
	type result struct {
		output string
		err    error
	}

	resultCh := make(chan result, 1)

	go func() {
		// wait for the container to finish
		statusCh, errCh := svc.cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

		select {
		case err := <-errCh:
			if err != nil {
				resultCh <- result{"", err}
				return
			}
		case <-statusCh:
		}

		r, err := svc.cli.ContainerLogs(context.Background(), containerID, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})

		if err != nil {
			resultCh <- result{"", err}
			return
		}
		defer r.Close()

		var buf bytes.Buffer
		if _, err := stdcopy.StdCopy(&buf, &buf, r); err != nil {
			resultCh <- result{"", err}
			return
		}

		resultCh <- result{buf.String(), nil}
	}()

	select {
	case res := <-resultCh:
		return res.output, res.err
	case <-execCtx.Done():
		return "", execCtx.Err() // 返回 context.DeadlineExceeded 或 context.Canceled
	}
}

func (svc *service) RunContainer(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (string, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return "", ErrNoWorkspaceID
	}

	if !isValidWorkspaceID(workspaceID) {
		return "", ErrWorkspaceNotFound
	}

	// sudo docker pull image
	_, err := svc.cli.ImageInspect(ctx, image)
	if err != nil {
		_, err := svc.cli.ImagePull(ctx, image, client.ImagePullOptions{})
		if err != nil {
			return "", err
		}
	}

	binds := []string{
		"/etc/localtime:/etc/localtime:ro",
	}

	if mountPath != "" {
		volumeName, err := svc.workspaceVolume(ctx, workspaceID)
		if err != nil {
			return "", err
		}

		binds = append(binds, volumeName+":"+mountPath)
	}

	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	networkName := svc.networkName
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}

	config := &container.Config{
		Image:     image,
		Cmd:       cmd,
		OpenStdin: true,
		Tty:       false,
		Labels: map[string]string{
			"workspace": workspaceID,
		},
	}

	if workDir != "" {
		config.WorkingDir = workDir
	}

	// sudo docker run -id -v <workspace_dir>:/workspace <image> /bin/sh
	resp, err := svc.cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, "")

	if err != nil {
		return "", err
	}

	containerID := resp.ID

	if err := svc.cli.ContainerStart(ctx, containerID,
		client.ContainerStartOptions{},
	); err != nil {
		return "", err
	}

	hijack, err := svc.cli.ContainerAttach(ctx, containerID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})

	if err != nil {
		return "", err
	}

	shortID := containerID[:12]

	svc.Lock()
	svc.sessions[shortID] = &hijack
	svc.Unlock()

	return containerID, nil
}

func (svc *service) reconnectToContainer(ctx context.Context, containerID string) error {
	inspect, err := svc.cli.ContainerInspect(ctx, containerID) // check if the container exists
	if err != nil {
		return err
	}

	if !inspect.State.Running {
		return ErrContainerNotRunning
	}

	hijack, err := svc.cli.ContainerAttach(ctx, containerID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})

	if err != nil {
		return err
	}

	shortID := containerID
	if len(containerID) > 12 {
		shortID = containerID[:12]
	}

	svc.Lock()
	svc.sessions[shortID] = &hijack
	svc.Unlock()

	return nil
}

func (svc *service) isContainerInWorkspace(ctx context.Context, containerID string, workspaceID string) (bool, error) {
	inspect, err := svc.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}

	return inspect.Config.Labels["workspace"] == workspaceID, nil
}

func (svc *service) SendToContainer(ctx context.Context, containerID string, input string) error {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return ErrNoWorkspaceID
	}

	ok, err := svc.isContainerInWorkspace(ctx, containerID, workspaceID)
	if err != nil {
		return err
	}

	if !ok {
		return ErrContainerNotFound
	}

	switch input {
	case "<ctrl+c>":
		// Send SIGINT to all processes except init
		_, err := svc.ExecCommand(ctx, containerID, "sh", "-c", "kill -INT -1")
		return err

	case "<ctrl+z>":
		// Send SIGTSTP to all processes except init
		_, err := svc.ExecCommand(ctx, containerID, "sh", "-c", "kill -TSTP -1")
		return err
	}

	shortID := containerID
	if len(containerID) > 12 {
		shortID = containerID[:12]
	}

	svc.RLock()
	hijack, ok := svc.sessions[shortID]
	svc.RUnlock()

	// Try to reconnect if session not found
	if !ok {
		if err := svc.reconnectToContainer(ctx, containerID); err != nil {
			return err
		}

		svc.RLock()
		hijack, ok = svc.sessions[shortID]
		svc.RUnlock()

		if !ok {
			return ErrContainerNotFound
		}
	}

	// Send input to container's stdin with newline
	writer := bufio.NewWriter(hijack.Conn)
	writer.WriteString(input + "\n")
	writer.Flush()

	return nil
}

func (svc *service) LogsContainer(ctx context.Context, containerID string, since time.Time, tail ...int) (string, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return "", ErrNoWorkspaceID
	}

	ok, err := svc.isContainerInWorkspace(ctx, containerID, workspaceID)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", ErrContainerNotFound
	}

	opts := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	if !since.IsZero() {
		opts.Since = since.Format(time.RFC3339)
	}

	if len(tail) > 0 && tail[0] > 0 {
		opts.Tail = strconv.Itoa(tail[0])
	}

	// sudo docker logs --tail <tail> <containerID>
	r, err := svc.cli.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var buf bytes.Buffer
	if _, err := stdcopy.StdCopy(&buf, &buf, r); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (svc *service) ExecCommand(ctx context.Context, containerID string, cmd ...string) (string, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return "", ErrNoWorkspaceID
	}

	ok, err := svc.isContainerInWorkspace(ctx, containerID, workspaceID)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", ErrContainerNotFound
	}

	execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// sudo docker exec <containerID> <command>
	resp, err := svc.cli.ContainerExecCreate(execCtx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	})

	if err != nil {
		return "", err
	}

	hijack, err := svc.cli.ContainerExecAttach(execCtx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", err
	}
	defer hijack.Close()

	// 使用 channel 來處理讀取操作的逾時
	type result struct {
		output string
		err    error
	}

	resultCh := make(chan result, 1)

	go func() {
		var buf bytes.Buffer
		_, err := stdcopy.StdCopy(&buf, &buf, hijack.Reader)

		resultCh <- result{
			output: buf.String(),
			err:    err,
		}
	}()

	select {
	case res := <-resultCh:
		return res.output, res.err
	case <-execCtx.Done():
		return "", execCtx.Err() // 返回 context.DeadlineExceeded 或 context.Canceled
	}
}

func (svc *service) Wait(ctx context.Context, timeout time.Duration) {
	time.Sleep(timeout)
}

func (svc *service) SendAndRead(ctx context.Context, containerID string, input string, wait time.Duration) (string, error) {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return "", ErrNoWorkspaceID
	}

	ok, err := svc.isContainerInWorkspace(ctx, containerID, workspaceID)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", ErrContainerNotFound
	}

	shortID := containerID
	if len(containerID) > 12 {
		shortID = containerID[:12]
	}

	svc.RLock()
	hijack, ok := svc.sessions[shortID]
	svc.RUnlock()

	if !ok {
		return "", ErrContainerNotFound
	}

	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	var buf bytes.Buffer
	go func(ctx context.Context, buf *bytes.Buffer) {
		tmp := make([]byte, 4096)
		deadline := time.Now().Add(wait)
		for time.Now().Before(deadline) {
			hijack.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := hijack.Reader.Read(tmp)
			if n > 0 {
				stdcopy.StdCopy(buf, buf, bytes.NewReader(tmp[:n]))
			}

			if err != nil && !os.IsTimeout(err) {
				break
			}

			time.Sleep(30 * time.Millisecond)
		}
	}(ctx, &buf)

	writer := bufio.NewWriter(hijack.Conn)
	writer.WriteString(input + "\n")
	writer.Flush()

	<-ctx.Done()

	return buf.String(), nil
}

func (svc *service) RemoveContainer(ctx context.Context, containerID string) error {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return ErrNoWorkspaceID
	}

	ok, err := svc.isContainerInWorkspace(ctx, containerID, workspaceID)
	if err != nil {
		return err
	}

	if !ok {
		return ErrContainerNotFound
	}

	shortID := containerID
	if len(containerID) > 12 {
		shortID = containerID[:12]
	}

	svc.Lock()
	if hijack, ok := svc.sessions[shortID]; ok {
		hijack.Close()
		delete(svc.sessions, shortID)
	}
	svc.Unlock()

	// sudo docker rm -f <containerID>
	return svc.cli.ContainerRemove(ctx, containerID,
		client.ContainerRemoveOptions{Force: true},
	)
}

func (svc *service) RemoveAllContainers(ctx context.Context) error {
	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if !ok {
		return ErrNoWorkspaceID
	}

	filters := filters.NewArgs()
	filters.Add("label", "workspace="+workspaceID)

	containers, err := svc.cli.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: filters,
	})

	if err != nil {
		return err
	}

	for _, c := range containers {
		shortID := c.ID
		if len(c.ID) > 12 {
			shortID = c.ID[:12]
		}

		svc.Lock()
		if hijack, ok := svc.sessions[shortID]; ok {
			hijack.Close()
			delete(svc.sessions, shortID)
		}
		svc.Unlock()

		// sudo docker rm -f <containerID>
		if err := svc.cli.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true}); err != nil {
			return err
		}
	}

	return nil
}

func (svc *service) Close(ctx context.Context) error {
	svc.Lock()
	defer svc.Unlock()

	for id, hijack := range svc.sessions {
		hijack.Close()
		delete(svc.sessions, id)
	}

	return svc.cli.Close()
}

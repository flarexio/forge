package forge

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

func ImageDescriptionEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		image, ok := request.(string)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.ImageDescription(ctx, image)
	}
}

type ListImagesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func ListImagesEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(ListImagesRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.ListImages(ctx, req.Page, req.PageSize)
	}
}

func PullImageEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		image, ok := request.(string)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return nil, svc.PullImage(ctx, image)
	}
}

func ListContainersEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		return svc.ListContainers(ctx)
	}
}

type RunContainerRequest struct {
	Image     string   `json:"image"`
	MountPath string   `json:"mount_path"`
	WorkDir   string   `json:"workdir"`
	Cmd       []string `json:"cmd"`
}

func RunContainerOnceEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(RunContainerRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.RunContainerOnce(ctx, req.Image, req.MountPath, req.WorkDir, req.Cmd...)
	}
}

func RunContainerEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(RunContainerRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.RunContainer(ctx, req.Image, req.MountPath, req.WorkDir, req.Cmd...)
	}
}

type SendToContainerRequest struct {
	ContainerID string `json:"container_id"`
	Input       string `json:"input"`
}

func SendToContainerEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SendToContainerRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return nil, svc.SendToContainer(ctx, req.ContainerID, req.Input)
	}
}

type LogsContainerRequest struct {
	ContainerID string
	Since       time.Time
	Tail        int
}

func (req *LogsContainerRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		ContainerID string `json:"container_id"`
		Since       string `json:"since"`
		Tail        int    `json:"tail"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	req.ContainerID = raw.ContainerID
	req.Tail = raw.Tail

	req.Since = time.Time{}
	if raw.Since != "" {
		since, err := time.ParseInLocation(time.RFC3339, raw.Since, time.Local)
		if err != nil {
			return err
		}

		req.Since = since
	}

	return nil
}

func LogsContainerEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(LogsContainerRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.LogsContainer(ctx, req.ContainerID, req.Since, req.Tail)
	}
}

type ExecCommandRequest struct {
	ContainerID string   `json:"container_id"`
	Cmd         []string `json:"cmd"`
}

func ExecCommandEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(ExecCommandRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.ExecCommand(ctx, req.ContainerID, req.Cmd...)
	}
}

type WaitRequest struct {
	Timeout time.Duration `json:"timeout"`
}

func WaitEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(WaitRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		svc.Wait(ctx, req.Timeout)
		return nil, nil
	}
}

type SendAndReadRequest struct {
	ContainerID string
	Input       string
	Wait        time.Duration
}

func (req *SendAndReadRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		ContainerID string `json:"container_id"`
		Input       string `json:"input"`
		Wait        string `json:"wait"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	req.ContainerID = raw.ContainerID
	req.Input = raw.Input

	req.Wait = 5 * time.Second
	if raw.Wait != "" {
		wait, err := time.ParseDuration(raw.Wait)
		if err != nil {
			return err
		}

		req.Wait = wait
	}

	return nil
}

func SendAndReadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SendAndReadRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return svc.SendAndRead(ctx, req.ContainerID, req.Input, req.Wait)
	}
}

type RemoveContainerRequest struct {
	ContainerID string `json:"container_id"`
}

func RemoveContainerEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(RemoveContainerRequest)
		if !ok {
			return nil, ErrInvalidRequest
		}

		return nil, svc.RemoveContainer(ctx, req.ContainerID)
	}
}

func RemoveAllContainersEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		return nil, svc.RemoveAllContainers(ctx)
	}
}

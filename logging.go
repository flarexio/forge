package forge

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(log *zap.Logger) ServiceMiddleware {
	return func(next Service) Service {
		log := log.With(
			zap.String("service", "forge"),
		)

		log.Info("service running")

		return &loggingMiddleware{
			log:  log,
			next: next,
		}
	}
}

type loggingMiddleware struct {
	log  *zap.Logger
	next Service
}

func (mw *loggingMiddleware) ImageDescription(ctx context.Context, image string) (string, error) {
	log := mw.log.With(
		zap.String("action", "image_description"),
		zap.String("image", image),
	)

	description, err := mw.next.ImageDescription(ctx, image)
	if err != nil {
		log.Error("failed to get image description", zap.Error(err))
		return "", err
	}

	log.Info("got image description")
	return description, nil
}

func (mw *loggingMiddleware) ListImages(ctx context.Context, page int, pageSize int) (images []string, err error) {
	log := mw.log.With(
		zap.String("action", "list_images"),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	images, err = mw.next.ListImages(ctx, page, pageSize)
	if err != nil {
		log.Error("failed to list images", zap.Error(err))
		return nil, err
	}

	log.Info("listed images", zap.Int("count", len(images)))
	return images, nil
}

func (mw *loggingMiddleware) PullImage(ctx context.Context, image string) error {
	log := mw.log.With(
		zap.String("action", "pull_image"),
		zap.String("image", image),
	)

	start := time.Now()
	err := mw.next.PullImage(ctx, image)
	if err != nil {
		log.Error("failed to pull image", zap.Error(err))
		return err
	}

	log.Info("pulled image", zap.Duration("took", time.Since(start)))
	return nil
}

func (mw *loggingMiddleware) ListContainers(ctx context.Context) ([]*Container, error) {
	log := mw.log.With(
		zap.String("action", "list_containers"),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	containers, err := mw.next.ListContainers(ctx)
	if err != nil {
		log.Error("failed to list containers", zap.Error(err))
		return nil, err
	}

	log.Info("listed containers", zap.Int("count", len(containers)))
	return containers, nil
}

func (mw *loggingMiddleware) RunContainerOnce(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (result string, err error) {
	log := mw.log.With(
		zap.String("action", "run_container_once"),
		zap.String("image", image),
		zap.Strings("cmd", cmd),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	start := time.Now()
	result, err = mw.next.RunContainerOnce(ctx, image, mountPath, workDir, cmd...)
	if err != nil {
		log.Error("failed to run container once", zap.Error(err))
		return "", err
	}

	log.Info("ran container once", zap.Duration("took", time.Since(start)))
	return result, nil
}

func (mw *loggingMiddleware) RunContainer(ctx context.Context, image string, mountPath string, workDir string, cmd ...string) (containerID string, err error) {
	log := mw.log.With(
		zap.String("action", "run_container"),
		zap.String("image", image),
		zap.Strings("cmd", cmd),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	start := time.Now()
	containerID, err = mw.next.RunContainer(ctx, image, mountPath, workDir, cmd...)
	if err != nil {
		log.Error("failed to run container", zap.Error(err))
		return "", err
	}

	log.Info("ran container", zap.String("container", containerID), zap.Duration("took", time.Since(start)))
	return containerID, nil
}

func (mw *loggingMiddleware) SendToContainer(ctx context.Context, containerID string, input string) error {
	log := mw.log.With(
		zap.String("action", "send_to_container"),
		zap.String("container", containerID),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	err := mw.next.SendToContainer(ctx, containerID, input)
	if err != nil {
		log.Error("failed to send to container", zap.Error(err))
		return err
	}

	log.Info("sent to container")
	return nil
}

func (mw *loggingMiddleware) LogsContainer(ctx context.Context, containerID string, since time.Time, tail ...int) (output string, err error) {
	log := mw.log.With(
		zap.String("action", "logs_container"),
		zap.String("container", containerID),
		zap.Time("since", since),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	output, err = mw.next.LogsContainer(ctx, containerID, since, tail...)
	if err != nil {
		log.Error("failed to get logs from container", zap.Error(err))
		return "", err
	}

	log.Info("got logs from container")
	return output, nil
}

func (mw *loggingMiddleware) ExecCommand(ctx context.Context, containerID string, cmd ...string) (output string, err error) {
	log := mw.log.With(
		zap.String("action", "exec_command"),
		zap.String("container", containerID),
		zap.Strings("cmd", cmd),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	start := time.Now()
	output, err = mw.next.ExecCommand(ctx, containerID, cmd...)
	if err != nil {
		log.Error("failed to exec command in container", zap.Error(err))
		return "", err
	}

	log.Info("executed command in container", zap.Duration("took", time.Since(start)))
	return output, nil
}

func (mw *loggingMiddleware) Wait(ctx context.Context, duration time.Duration) {
	log := mw.log.With(
		zap.String("action", "wait"),
		zap.Duration("duration", duration),
	)

	mw.next.Wait(ctx, duration)
	log.Info("waited")
}

func (mw *loggingMiddleware) SendAndRead(ctx context.Context, containerID string, input string, wait time.Duration) (output string, err error) {
	log := mw.log.With(
		zap.String("action", "send_and_read"),
		zap.String("container", containerID),
		zap.String("input", input),
		zap.Duration("wait", wait),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	start := time.Now()
	output, err = mw.next.SendAndRead(ctx, containerID, input, wait)
	if err != nil {
		log.Error("failed to send and read from container", zap.Error(err))
		return "", err
	}

	log.Info("sent and read from container", zap.Duration("took", time.Since(start)))
	return output, nil
}

func (mw *loggingMiddleware) RemoveContainer(ctx context.Context, containerID string) error {
	log := mw.log.With(
		zap.String("action", "remove_container"),
		zap.String("container", containerID),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	err := mw.next.RemoveContainer(ctx, containerID)
	if err != nil {
		log.Error("failed to remove container", zap.Error(err))
		return err
	}

	log.Info("removed container")
	return nil
}

func (mw *loggingMiddleware) RemoveAllContainers(ctx context.Context) error {
	log := mw.log.With(
		zap.String("action", "remove_all_containers"),
	)

	workspaceID, ok := ctx.Value(WorkspaceID).(string)
	if ok {
		log = log.With(zap.String("workspace", workspaceID))
	}

	err := mw.next.RemoveAllContainers(ctx)
	if err != nil {
		log.Error("failed to remove all containers", zap.Error(err))
		return err
	}

	log.Info("removed all containers")
	return nil
}

func (mw *loggingMiddleware) Close(ctx context.Context) error {
	log := mw.log.With(
		zap.String("action", "close"),
	)

	err := mw.next.Close(ctx)
	if err != nil {
		log.Error("failed to close service", zap.Error(err))
		return err
	}

	log.Info("service closed")
	return nil
}

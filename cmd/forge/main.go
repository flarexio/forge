package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"github.com/flarexio/forge"
	"github.com/flarexio/forge/transport/http"
)

func main() {
	cmd := &cli.Command{
		Name:  "forge",
		Usage: "A lightweight container management tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Specifies the working directory",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "HTTP server port",
				Value: 8080,
			},
		},
		Action: run,
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	path := cmd.String("path")
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path = filepath.Join(homeDir, ".flarex", "forge")
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	svc, err := forge.NewService(path)
	if err != nil {
		return err
	}

	svc = forge.LoggingMiddleware(log)(svc)

	defer svc.Close(ctx)

	r := gin.Default()

	// GET /docker/images/:image/desc
	// GET /docker/:namespace/images/:image/desc
	{
		endpoint := forge.ImageDescriptionEndpoint(svc)
		handler := http.ImageDescriptionHandler(endpoint)
		r.GET("/docker/images/:image/desc", handler)
		r.GET("/docker/:namespace/images/:image/desc", handler)
	}

	// GET /docker/images
	{
		endpoint := forge.ListImagesEndpoint(svc)
		handler := http.ListImagesHandler(endpoint)
		r.GET("/docker/images", handler)
	}

	// GET /docker/images/:image/pull
	// GET /docker/:namespace/images/:image/pull
	{
		endpoint := forge.PullImageEndpoint(svc)
		handler := http.PullImageHandler(endpoint)
		r.GET("/docker/images/:image/pull", handler)
		r.GET("/docker/:namespace/images/:image/pull", handler)
	}

	// GET /workspaces/:workspace/containers
	{
		endpoint := forge.ListContainersEndpoint(svc)
		handler := http.ListContainersHandler(endpoint)
		r.GET("/workspaces/:workspace/containers", handler)
	}

	// POST /workspaces/:workspace/containers/run_once
	{
		endpoint := forge.RunContainerOnceEndpoint(svc)
		handler := http.RunContainerOnceHandler(endpoint)
		r.POST("/workspaces/:workspace/containers/run_once", handler)
	}

	// POST /workspaces/:workspace/containers/run
	{
		endpoint := forge.RunContainerEndpoint(svc)
		handler := http.RunContainerHandler(endpoint)
		r.POST("/workspaces/:workspace/containers/run", handler)
	}

	// POST /workspaces/:workspace/containers/:container/send
	{
		endpoint := forge.SendToContainerEndpoint(svc)
		handler := http.SendToContainerHandler(endpoint)
		r.POST("/workspaces/:workspace/containers/:container/send", handler)
	}

	// GET /workspaces/:workspace/containers/:container/logs
	{
		endpoint := forge.LogsContainerEndpoint(svc)
		handler := http.LogsContainerHandler(endpoint)
		r.GET("/workspaces/:workspace/containers/:container/logs", handler)
	}

	// POST /workspaces/:workspace/containers/:container/exec
	{
		endpoint := forge.ExecCommandEndpoint(svc)
		handler := http.ExecCommandHandler(endpoint)
		r.POST("/workspaces/:workspace/containers/:container/exec", handler)
	}

	// GET /workspaces/:workspace/wait
	{
		endpoint := forge.WaitEndpoint(svc)
		handler := http.WaitHandler(endpoint)
		r.GET("/workspaces/:workspace/wait", handler)
	}

	// POST /workspaces/:workspace/containers/:container/send_and_read
	{
		endpoint := forge.SendAndReadEndpoint(svc)
		handler := http.SendAndReadHandler(endpoint)
		r.POST("/workspaces/:workspace/containers/:container/send_and_read", handler)
	}

	// DELETE /workspaces/:workspace/containers/:container
	{
		endpoint := forge.RemoveContainerEndpoint(svc)
		handler := http.RemoveContainerHandler(endpoint)
		r.DELETE("/workspaces/:workspace/containers/:container", handler)
	}

	// DELETE /workspaces/:workspace/containers
	{
		endpoint := forge.RemoveAllContainersEndpoint(svc)
		handler := http.RemoveAllContainersHandler(endpoint)
		r.DELETE("/workspaces/:workspace/containers", handler)
	}

	go r.Run(":" + strconv.Itoa(cmd.Int("port")))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit

	log.Info("graceful shutdown", zap.String("signal", sign.String()))
	return nil
}

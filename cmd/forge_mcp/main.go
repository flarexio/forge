package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/urfave/cli/v3"

	"github.com/flarexio/forge"
	"github.com/flarexio/forge/transport/mcp"
)

func main() {
	cmd := cli.Command{
		Name:  "forge_mcp",
		Usage: "A lightweight container management tool - MCP Server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Specifies the working directory",
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

	svc, err := forge.NewService(path)
	if err != nil {
		return err
	}
	defer svc.Close(ctx)

	s := server.NewMCPServer(
		"Forge MCP Server",
		"1.0.0",
		server.WithToolHandlerMiddleware(mcp.InjectContextMiddleware()),
	)

	// Add ImageDescription tool
	{
		endpoint := forge.ImageDescriptionEndpoint(svc)
		handler := mcp.ImageDescriptionHandler(endpoint)
		tool := mcp.ImageDescriptionTool()
		s.AddTool(tool, handler)
	}

	// Add ListImages tool
	{
		endpoint := forge.ListImagesEndpoint(svc)
		handler := mcp.ListImagesHandler(endpoint)
		tool := mcp.ListImagesTool()
		s.AddTool(tool, handler)
	}

	// Add PullImage tool
	{
		endpoint := forge.PullImageEndpoint(svc)
		handler := mcp.PullImageHandler(endpoint)
		tool := mcp.PullImageTool()
		s.AddTool(tool, handler)
	}

	// Add ListContainers tool
	{
		endpoint := forge.ListContainersEndpoint(svc)
		handler := mcp.ListContainersHandler(endpoint)
		tool := mcp.ListContainersTool()
		s.AddTool(tool, handler)
	}

	// Add RunContainerOnce tool
	{
		endpoint := forge.RunContainerOnceEndpoint(svc)
		handler := mcp.RunContainerOnceHandler(endpoint)
		tool := mcp.RunContainerOnceTool()
		s.AddTool(tool, handler)
	}

	// Add RunContainer tool
	{
		endpoint := forge.RunContainerEndpoint(svc)
		handler := mcp.RunContainerHandler(endpoint)
		tool := mcp.RunContainerTool()
		s.AddTool(tool, handler)
	}

	// Add SendToContainer tool
	{
		endpoint := forge.SendToContainerEndpoint(svc)
		handler := mcp.SendToContainerHandler(endpoint)
		tool := mcp.SendToContainerTool()
		s.AddTool(tool, handler)
	}

	// Add LogsContainer tool
	{
		endpoint := forge.LogsContainerEndpoint(svc)
		handler := mcp.LogsContainerHandler(endpoint)
		tool := mcp.LogsContainerTool()
		s.AddTool(tool, handler)
	}

	// Add ExecCommand tool
	{
		endpoint := forge.ExecCommandEndpoint(svc)
		handler := mcp.ExecCommandHandler(endpoint)
		tool := mcp.ExecCommandTool()
		s.AddTool(tool, handler)
	}

	// Add Wait tool
	{
		endpoint := forge.WaitEndpoint(svc)
		handler := mcp.WaitHandler(endpoint)
		tool := mcp.WaitTool()
		s.AddTool(tool, handler)
	}

	// Add SendAndRead tool
	{
		endpoint := forge.SendAndReadEndpoint(svc)
		handler := mcp.SendAndReadHandler(endpoint)
		tool := mcp.SendAndReadTool()
		s.AddTool(tool, handler)
	}

	// Add RemoveContainer tool
	{
		endpoint := forge.RemoveContainerEndpoint(svc)
		handler := mcp.RemoveContainerHandler(endpoint)
		tool := mcp.RemoveContainerTool()
		s.AddTool(tool, handler)
	}

	// Add RemoveAllContainers tool
	{
		endpoint := forge.RemoveAllContainersEndpoint(svc)
		handler := mcp.RemoveAllContainersHandler(endpoint)
		tool := mcp.RemoveAllContainersTool()
		s.AddTool(tool, handler)
	}

	go server.ServeStdio(s)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	return nil
}

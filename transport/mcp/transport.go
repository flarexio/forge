package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flarexio/forge"
)

func ImageDescriptionTool(name ...string) mcp.Tool {
	toolName := "ImageDescription"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Get the description of an image"),
		mcp.WithString("image",
			mcp.Description("The image to describe"),
			mcp.Required(),
		),
	)
}

func ImageDescriptionHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		image := request.GetString("image", "")
		if image == "" {
			return mcp.NewToolResultError("image is required"), nil
		}

		resp, err := endpoint(ctx, image)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		desc, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(desc), nil
	}
}

func ListImagesTool(name ...string) mcp.Tool {
	toolName := "ListImages"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("List all images"),
		mcp.WithNumber("page",
			mcp.Description("The page number (starting from 1)"),
			mcp.DefaultNumber(1),
		),
		mcp.WithNumber("page_size",
			mcp.Description("The number of images per page"),
			mcp.DefaultNumber(20),
		),
	)
}

func ListImagesHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.ListImagesRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		images, ok := resp.([]string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		bs, err := json.Marshal(&images)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(bs)), nil
	}
}

func PullImageTool(name ...string) mcp.Tool {
	toolName := "PullImage"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Pull an image"),
		mcp.WithString("image",
			mcp.Description("The image to pull"),
			mcp.Required(),
		),
	)
}

func PullImageHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		image := request.GetString("image", "")
		if image == "" {
			return mcp.NewToolResultError("image is required"), nil
		}

		_, err := endpoint(ctx, image)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("ok"), nil
	}
}

func ListContainersTool(name ...string) mcp.Tool {
	toolName := "ListContainers"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("List all containers in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
	)
}

func ListContainersHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := endpoint(ctx, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		bs, err := json.Marshal(&resp)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(bs)), nil
	}
}

func RunContainerOnceTool(name ...string) mcp.Tool {
	toolName := "RunContainerOnce"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Run a container once in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("image",
			mcp.Description("The image to run"),
			mcp.Required(),
		),
		mcp.WithString("mount_path",
			mcp.Description("The path to mount the workspace (if empty, no mount)"),
		),
		mcp.WithString("workdir",
			mcp.Description("The working directory inside the container (if empty, default)"),
		),
		mcp.WithArray("cmd",
			mcp.Description("The command to run"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	)
}

func RunContainerOnceHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.RunContainerRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func RunContainerTool(name ...string) mcp.Tool {
	toolName := "RunContainer"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Run a container in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("image",
			mcp.Description("The image to run"),
			mcp.Required(),
		),
		mcp.WithString("mount_path",
			mcp.Description("The path to mount the workspace (if empty, no mount)"),
		),
		mcp.WithString("workdir",
			mcp.Description("The working directory inside the container (if empty, default)"),
		),
		mcp.WithArray("cmd",
			mcp.Description("The command to run"),
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Required(),
		),
	)
}

func RunContainerHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.RunContainerRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		containerID, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(containerID), nil
	}
}

func SendToContainerTool(name ...string) mcp.Tool {
	toolName := "SendToContainer"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Send a command to a running container in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string",
				mcp.Description("The ID of the workspace"),
			),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container"),
			mcp.Required(),
		),
		mcp.WithString("input",
			mcp.Description("The command to run"),
			mcp.Required(),
		),
	)
}

func SendToContainerHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.SendToContainerRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		startTime := time.Now()

		_, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := "ok (sent at " + startTime.Format(time.RFC3339) + ")"

		return mcp.NewToolResultText(result), nil
	}
}

func LogsContainerTool(name ...string) mcp.Tool {
	toolName := "LogsContainer"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Get the logs of a container in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container"),
			mcp.Required(),
		),
		mcp.WithString("since",
			mcp.Description("Show logs since this timestamp (RFC3339)"),
			mcp.Required(),
		),
		mcp.WithNumber("tail",
			mcp.Description("The number of lines to show from the end of the logs (0 or not set means all lines)"),
		),
	)
}

func LogsContainerHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.LogsContainerRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		logs, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(logs), nil
	}
}

func ExecCommandTool(name ...string) mcp.Tool {
	toolName := "ExecCommand"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Execute a command in a running container in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container"),
			mcp.Required(),
		),
		mcp.WithArray("cmd",
			mcp.Description("The command to run"),
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Required(),
		),
	)
}

func ExecCommandHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.ExecCommandRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func WaitTool(name ...string) mcp.Tool {
	toolName := "Wait"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Wait for a container to finish in the current workspace"),
		mcp.WithString("timeout",
			mcp.Description("The maximum time to wait for the container to finish (Go duration format, e.g. '5s', '10s', '1m')"),
			mcp.Required(),
		),
	)
}

func WaitHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		timeoutStr := request.GetString("timeout", "10s")

		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return mcp.NewToolResultError("invalid timeout format"), nil
		}

		req := forge.WaitRequest{
			Timeout: timeout,
		}

		endpoint(ctx, req)

		return mcp.NewToolResultText("ok"), nil
	}
}

func SendAndReadTool(name ...string) mcp.Tool {
	toolName := "SendAndRead"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Send input to a running container and read the output in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string",
				mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container"),
			mcp.Required(),
		),
		mcp.WithString("input",
			mcp.Description("The input to send to the container"),
			mcp.Required(),
		),
		mcp.WithString("wait",
			mcp.Description("The maximum time to wait for a response (Go duration format, e.g. '5s', '10s', '1m')"),
		),
	)
}

func SendAndReadHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.SendAndReadRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		output, ok := resp.(string)
		if !ok {
			err := errors.New("invalid response from service")
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}

func RemoveContainerTool(name ...string) mcp.Tool {
	toolName := "RemoveContainer"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Remove a container in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
		mcp.WithString("container_id",
			mcp.Description("The ID of the container"),
			mcp.Required(),
		),
	)
}

func RemoveContainerHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var req forge.RemoveContainerRequest
		if err := request.BindArguments(&req); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		_, err := endpoint(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("ok"), nil
	}
}

func RemoveAllContainersTool(name ...string) mcp.Tool {
	toolName := "RemoveAllContainers"
	if len(name) > 0 {
		toolName = name[0]
	}

	return mcp.NewTool(toolName,
		mcp.WithDescription("Remove all containers in the current workspace"),
		WithContext("ctx", "The context of the request",
			NewProperty("workspace_id", "string", mcp.Description("The ID of the workspace")),
		),
	)
}

func RemoveAllContainersHandler(endpoint endpoint.Endpoint) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_, err := endpoint(ctx, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText("ok"), nil
	}
}

package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/flarexio/forge"
)

type Property struct {
	name   string
	schema map[string]any
}

func NewProperty(name string, dataType string, opts ...mcp.PropertyOption) *Property {
	prop := &Property{
		name: name,
		schema: map[string]any{
			"type": dataType,
		},
	}

	for _, opt := range opts {
		opt(prop.schema)
	}

	delete(prop.schema, "required")

	return prop
}

func WithContext(name string, desc string, props ...*Property) mcp.ToolOption {
	properties := make(map[string]any)
	for _, prop := range props {
		properties[prop.name] = prop.schema
	}

	return mcp.WithObject(name,
		mcp.Description(desc),
		mcp.Properties(properties),
		mcp.Required(),
	)
}

type ForgeContext struct {
	WorkspaceID string `json:"workspace_id"`
}

func InjectContextMiddleware() server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var req struct {
				ForgeCtx ForgeContext `json:"ctx"`
			}

			if err := request.BindArguments(&req); err != nil {
				return next(ctx, request)
			}

			if req.ForgeCtx.WorkspaceID != "" {
				ctx = context.WithValue(ctx, forge.WorkspaceID, req.ForgeCtx.WorkspaceID)
			}

			return next(ctx, request)
		}
	}
}

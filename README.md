# Forge - Lightweight Container Management Tool

A modern Docker container management service that provides both HTTP API and MCP (Model Context Protocol) server for seamless AI agent integration.

## Features

- 🐳 **Docker Container Management**: Create, run, and manage Docker containers with workspace isolation
- 🤖 **AI Agent Ready**: Built-in MCP server support for AI agents and assistants like Claude
- 🔒 **Workspace Isolation**: Each workspace gets its own Docker volume for security
- 📡 **Dual Interface**: Both HTTP REST API and MCP protocol support
- 🚀 **Lightweight**: Minimal overhead with clean, focused functionality
- 🔧 **Developer Friendly**: Simple CLI interface with sensible defaults
- 🌐 **Network Isolation**: Dedicated Docker network for forge containers

## Use Cases

- AI-powered development environments
- Isolated code execution for AI assistants
- Container-based CI/CD workflows
- Secure multi-tenant development platforms
- Interactive programming sessions with AI agents

## Installation

```bash
go install github.com/flarexio/forge/cmd/forge@latest
go install github.com/flarexio/forge/cmd/forge_mcp@latest
```

## Quick Start

### HTTP Server Mode

Start the HTTP API server:

```bash
forge --port 8080 --path /path/to/workspace
```

The server will be available at `http://localhost:8080`

### MCP Server Mode (for AI agents)

Start the MCP server for AI agent integration:

```bash
forge_mcp --path /path/to/workspace
```

## API Overview

### Container Operations

- **Run Once**: Execute a command in a container and remove it
- **Run Interactive**: Start a persistent container with shell access
- **Execute Commands**: Run commands in existing containers
- **Send Input**: Send input to container stdin (supports Ctrl+C, Ctrl+Z)
- **Read Logs**: Retrieve container logs with optional filters
- **Remove Containers**: Clean up containers individually or all at once

### Image Management

- **List Images**: Browse available Docker images
- **Pull Images**: Download images from registries
- **Image Description**: Get detailed information about images from Docker Hub

### Workspace Management

- **Workspace Isolation**: Each workspace ID gets its own Docker volume
- **Volume Mounting**: Optional mounting of workspace to container paths
- **Network Isolation**: Containers run in dedicated `forge` network

## HTTP API Examples

### List Available Images

```bash
curl "http://localhost:8080/docker/images?page=1&pageSize=10"
```

### Pull an Image

```bash
curl "http://localhost:8080/docker/images/alpine/pull"
```

### Run a One-time Command

```bash
curl -X POST "http://localhost:8080/workspaces/my-project/containers/run_once" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "node:18",
    "mountPath": "/workspace",
    "workDir": "/workspace",
    "cmd": ["npm", "install"]
  }'
```

### Start Interactive Container

```bash
curl -X POST "http://localhost:8080/workspaces/my-project/containers/run" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "python:3.9",
    "mountPath": "/app",
    "workDir": "/app",
    "cmd": ["python3"]
  }'
```

### Send Commands to Container

```bash
curl -X POST "http://localhost:8080/workspaces/my-project/containers/{container-id}/send" \
  -H "Content-Type: application/json" \
  -d '{
    "input": "print(\"Hello World\")"
  }'
```

### Execute Commands

```bash
curl -X POST "http://localhost:8080/workspaces/my-project/containers/{container-id}/exec" \
  -H "Content-Type: application/json" \
  -d '{
    "cmd": ["ls", "-la"]
  }'
```

## MCP Integration

Forge provides native MCP (Model Context Protocol) support for AI agents. The MCP server exposes all container management capabilities as tools that AI agents can use:

- `ImageDescription` - Get Docker image information from Docker Hub
- `ListImages` - Browse available Docker images with pagination
- `PullImage` - Download Docker images from registries
- `ListContainers` - Show all containers in the current workspace
- `RunContainerOnce` - Execute one-time commands in containers and remove them
- `RunContainer` - Start persistent interactive containers
- `SendToContainer` - Send input to running containers (supports Ctrl+C, Ctrl+Z)
- `LogsContainer` - Retrieve container logs with timestamp filtering
- `ExecCommand` - Execute commands in existing running containers
- `Wait` - Add configurable delays between operations
- `SendAndRead` - Send input to containers and read the output
- `RemoveContainer` - Remove specific containers by ID
- `RemoveAllContainers` - Clean up all containers in the workspace

### MCP Tool Parameters

Each tool requires appropriate parameters and context:

- **Context**: Most tools require `workspace_id` in the context
- **Container Operations**: Tools like `RunContainer` and `SendToContainer` need container configuration
- **Image Operations**: Tools like `PullImage` and `ImageDescription` work with image names
- **Timing**: Tools like `Wait` and `SendAndRead` support duration parameters

## Configuration

### Command Line Options

#### forge (HTTP Server)
```bash
forge [options]

Options:
  --path string   Specifies the working directory (default: ~/.flarex/forge)
  --port int      HTTP server port (default: 8080)
```

#### forge_mcp (MCP Server)
```bash
forge_mcp [options]

Options:
  --path string   Specifies the working directory (default: ~/.flarex/forge)
```

### Workspace Structure

```
~/.flarex/forge/
└── workspaces/
    ├── project-1/
    ├── project-2/
    └── my-workspace/
```

Each workspace directory is automatically mounted as a Docker volume for container access.

## Architecture

### Components

- **Service Layer**: Core business logic for container management
- **Transport Layer**: HTTP and MCP protocol handlers
- **Docker Integration**: Direct Docker API communication
- **Workspace Management**: Isolated workspace volumes
- **Network Management**: Dedicated container networking

### Security Features

- **Workspace ID Validation**: Prevents directory traversal attacks
- **Container Isolation**: Each workspace uses separate Docker volumes
- **Network Isolation**: Containers run in dedicated Docker network
- **Label-based Access Control**: Containers are tagged with workspace labels

## Development

### Prerequisites

- Go 1.21+
- Docker Engine
- Docker daemon running

### Build from Source

```bash
git clone https://github.com/flarexio/forge.git
cd forge

# Build HTTP server
go build -o bin/forge cmd/forge/main.go

# Build MCP server
go build -o bin/forge_mcp cmd/forge_mcp/main.go
```

### Run Tests

```bash
go test ./...
```

## Docker Requirements

- Docker Engine 20.10+
- Docker API version 1.41+
- Sufficient permissions to create containers, volumes, and networks

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Docker Engine API](https://docs.docker.com/engine/api/)
- MCP support via [mcp-go](https://github.com/mark3labs/mcp-go)
- HTTP server with [Gin](https://github.com/gin-gonic/gin)
- CLI with [urfave/cli](https://github.com/urfave/cli)
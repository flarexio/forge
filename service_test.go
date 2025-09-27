package forge

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type forgeServiceTestSuite struct {
	suite.Suite
	svc  Service
	path string
}

func (suite *forgeServiceTestSuite) SetupSuite() {
	path := "/tmp/forge_test"

	svc, err := NewService(path)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.svc = svc
	suite.path = path
}

func (suite *forgeServiceTestSuite) TestImageDescription() {
	ctx := context.Background()

	description, err := suite.svc.ImageDescription(ctx, "alpine")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.NotEmpty(description)
}

func (suite *forgeServiceTestSuite) TestListImages() {
	ctx := context.Background()

	images, err := suite.svc.ListImages(ctx, 1, 10)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Len(images, 10)
}

func (suite *forgeServiceTestSuite) TestRunContainerOnce() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, WorkspaceID, "workspace_1")

	// sudo docker run alpine:latest echo hello world
	output, err := suite.svc.RunContainerOnce(ctx, "alpine:latest", "", "", "echo", "hello world")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	output = strings.TrimSuffix(output, "\n")

	suite.Equal("hello world", output)
}

func (suite *forgeServiceTestSuite) TestRunContainerAndExecCommand() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, WorkspaceID, "workspace_2")

	// sudo docker run -id alpine:latest /bin/sh
	containerID, err := suite.svc.RunContainer(ctx, "alpine:latest", "", "", "/bin/sh")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	defer suite.svc.RemoveContainer(ctx, containerID)

	suite.NotEmpty(containerID)

	// sudo docker exec <container_id> echo "hello world"
	output, err := suite.svc.ExecCommand(ctx, containerID, []string{"echo", "hello world"}...)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	output = strings.TrimSuffix(output, "\n")

	suite.Equal("hello world", output)
}

func (suite *forgeServiceTestSuite) TestRunContainerWithInteractiveShell() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, WorkspaceID, "workspace_3")

	// sudo docker run -id alpine:latest /bin/sh
	containerID, err := suite.svc.RunContainer(ctx, "alpine:latest", "", "", "/bin/sh")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	defer suite.svc.RemoveContainer(ctx, containerID)

	suite.NotEmpty(containerID)

	conntainers, err := suite.svc.ListContainers(ctx)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	var ok bool
	for _, c := range conntainers {
		if c.ID == containerID {
			ok = true

			suite.Equal("alpine:latest", c.Image)
			suite.Equal("/bin/sh", c.Command)
		}
	}

	if !ok {
		suite.Fail("container not found")
		return
	}

	since := time.Now()
	if err := suite.svc.SendToContainer(ctx, containerID, "echo hello world\n"); err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.svc.Wait(ctx, 2*time.Second)

	output, err := suite.svc.LogsContainer(ctx, containerID, since, 1)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	output = strings.TrimSuffix(output, "\n")

	suite.Equal("hello world", output)
}

func (suite *forgeServiceTestSuite) TestInteractiveGolangRun() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, WorkspaceID, "workspace_4")

	containerID, err := suite.svc.RunContainer(ctx, "golang:1.24.5-alpine", "", "", "/bin/sh")
	if err != nil {
		suite.Fail(err.Error())
		return
	}
	defer suite.svc.RemoveContainer(ctx, containerID)

	suite.NotEmpty(containerID)

	since := time.Now()

	program := `package main

import "fmt"

func main() {
	fmt.Println("Hello, AI!")
}
`

	// cat > hello.go << EOF
	if err := suite.svc.SendToContainer(ctx, containerID, "cat > hello.go << EOF\n"+program+"EOF"); err != nil {
		suite.Fail(err.Error())
		return
	}

	// go run hello.go
	if err := suite.svc.SendToContainer(ctx, containerID, "go run hello.go\n"); err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.svc.Wait(ctx, 10*time.Second)

	output, err := suite.svc.LogsContainer(ctx, containerID, since, 10)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	output = strings.TrimSuffix(output, "\n")

	suite.Equal("Hello, AI!", output)
}

func (suite *forgeServiceTestSuite) TestSendAndRead() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, WorkspaceID, "workspace_5")

	containerID, err := suite.svc.RunContainer(ctx, "golang:1.24.5-alpine", "", "", "/bin/sh")
	if err != nil {
		suite.Fail(err.Error())
		return
	}
	defer suite.svc.RemoveContainer(ctx, containerID)

	suite.NotEmpty(containerID)

	program := `package main

import "fmt"

func main() {
	fmt.Println("Hello, AI!")
}
`

	// cat > hello.go << EOF
	if err := suite.svc.SendToContainer(ctx, containerID, "cat > hello.go << EOF\n"+program+"EOF"); err != nil {
		suite.Fail(err.Error())
		return
	}

	output, err := suite.svc.SendAndRead(ctx, containerID, "go run hello.go\n", 10*time.Second)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	output = strings.TrimSuffix(output, "\n")

	suite.Equal("Hello, AI!", output)
}

func (suite *forgeServiceTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.svc.Close(ctx)

	os.RemoveAll(suite.path)
}

func TestForgeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(forgeServiceTestSuite))
}

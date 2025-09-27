package forge

import (
	"encoding/json"
	"errors"
	"regexp"
)

type ContextKey string

const (
	WorkspaceID ContextKey = "workspace_id"
)

var (
	ErrNoWorkspaceID       = errors.New("no workspace ID in context")
	ErrWorkspaceNotFound   = errors.New("workspace not found")
	ErrContainerNotFound   = errors.New("container not found in workspace")
	ErrContainerNotRunning = errors.New("container is not running")
)

type Container struct {
	ID      string
	Image   string
	Command string
	Status  string
	Names   []string
}

func (c *Container) MarshalJSON() ([]byte, error) {
	shortID := c.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	output := struct {
		ID     string   `json:"id"`
		Image  string   `json:"image"`
		Status string   `json:"status"`
		Names  []string `json:"names"`
	}{
		ID:     shortID,
		Image:  c.Image,
		Status: c.Status,
		Names:  c.Names,
	}

	return json.Marshal(&output)
}

var re = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidWorkspaceID(id string) bool {
	if id == "" {
		return false
	}

	return re.MatchString(id)

}

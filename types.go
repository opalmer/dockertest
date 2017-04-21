package dockertest

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
)

var (
	// ErrPortNotFound is returned by ContainerInfo.Port if we're unable
	// to find a matching port on the c.
	ErrPortNotFound = errors.New("The requested port could not be found")
)

// ContainerInfo provides a wrapper around information
type ContainerInfo struct {
	JSON     types.ContainerJSON
	Data     types.Container
	State    *types.ContainerState
	Warnings []string
}

func (c *ContainerInfo) String() string {
	return fmt.Sprintf("ContainerInfo(image='%s', id='%s')", c.Data.Image, c.Data.ID)
}

// HasLabel returns true if the provided label exists and is equal
// to the provided value.
func (c *ContainerInfo) HasLabel(name string, value string) bool {
	currentValue, set := c.Data.Labels[name]
	return set && value == currentValue
}

// Port will return types.Port for the requested internal port.
func (c *ContainerInfo) Port(internal int) (types.Port, error) {
	for _, port := range c.Data.Ports {
		if port.PrivatePort == uint16(internal) {
			return port, nil
		}
	}
	return types.Port{}, ErrPortNotFound
}

// NewContainerInfo returns a *ContainerInfo struct.
func NewContainerInfo(container types.Container, json types.ContainerJSON) *ContainerInfo {
	return &ContainerInfo{
		Data: container, State: json.State, JSON: json,
		Warnings: []string{},
	}
}

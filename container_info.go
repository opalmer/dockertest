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
	return fmt.Sprintf("{image:%s, id:%s, status:%s}", c.Data.Image, c.Data.ID, c.Data.Status)
}

// GetLabel will return the value of the given label or "" if it does
// not exist. The boolean indicates if the label exists at all.
func (c *ContainerInfo) GetLabel(name string) (string, bool) {
	value, set := c.Data.Labels[name]
	return value, set
}

// HasLabel returns true if the provided label exists and is equal
// to the provided value.
func (c *ContainerInfo) HasLabel(name string, value string) bool {
	current, set := c.GetLabel(name)
	return set && value == current
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

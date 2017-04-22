package dockertest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
)

const timeNotSet = "0001-01-01T00:00:00Z"

var (
	// ErrPortNotFound is returned by ContainerInfo.Port if we're unable
	// to find a matching port on the c.
	ErrPortNotFound = errors.New("The requested port could not be found")

	// ErrContainerNotRunning is returned by Started() if the container
	// was never started.
	ErrContainerNotRunning = errors.New("container not running")

	// ErrContainerStillRunning is returned by Finished() if the container
	// is still running.
	ErrContainerStillRunning = errors.New("container still running")
)

// ContainerInfo provides a wrapper around information
type ContainerInfo struct {
	JSON     types.ContainerJSON
	Data     types.Container
	State    *types.ContainerState
	Warnings []string
	client   *DockerClient
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

// Refresh will refresh the data present on this struct.
func (c *ContainerInfo) Refresh() error {
	updated, err := c.client.ContainerInfo(context.Background(), c.ID())
	if err != nil {
		return err
	}
	*c = *updated
	return nil
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

// ID is a shortcut function to return the container's id
func (c *ContainerInfo) ID() string {
	return c.Data.ID
}

// Started returns the time the container was started at.
func (c *ContainerInfo) Started() (time.Time, error) {
	if c.State.StartedAt == timeNotSet {
		return time.Unix(0, 0).UTC(), ErrContainerNotRunning
	}
	return time.Parse(time.RFC3339Nano, c.State.StartedAt)
}

// Finished returns the time the container finished running.
func (c *ContainerInfo) Finished() (time.Time, error) {
	if c.State.FinishedAt == timeNotSet {
		return time.Unix(0, 0).UTC(), ErrContainerStillRunning
	}
	return time.Parse(time.RFC3339Nano, c.State.FinishedAt)
}

// Elapsed returns how long the container has been running or had run if
// the container has stopped.
func (c *ContainerInfo) Elapsed() (time.Duration, error) {
	started, err := c.Started()
	if err != nil {
		if err == ErrContainerNotRunning {
			return time.Second * 0, nil
		}
		return time.Second * 0, err
	}

	finished, err := c.Finished()
	if err != nil {
		if err == ErrContainerStillRunning {
			return time.Since(started), nil
		}
		return time.Second * 0, nil
	}
	return finished.Sub(started), nil
}

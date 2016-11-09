package dockertest

import (
	"fmt"
	"errors"

	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types"
)

// types.go provides a small set of basic types that are used and returned
// by dockertest.

var (
	// ErrPortNotFound is returned by Container.Port if we're unable
	// to find a matching port on the container.
	ErrPortNotFound = errors.New("")
)

// Container wraps the standard types.Container
type Container struct {
	types.Container
}

// NewContainer returns a *Container struct.
func NewContainer(container *types.Container) *Container {
	return &Container{container}
}

// HasLabel returns true if the provided label exists and is equal
// to the provided value.
func (container *Container) CheckLabel(name string, value string) bool {
	currentValue, set := container.Labels[name]
	return set && value == currentValue
}

// Port will return types.Port for the requested internal port.
func (container *Container) Port(internal int) (types.Port, error) {
	for _, port := range container.Ports {
		if port.PrivatePort == uint16(internal) {
			return port, nil
		}
	}
	return types.Port{}, ErrPortNotFound
}

// Ports is when to convey port exposures to RunContainer()
type Ports struct {
	specs []string
	publishall bool
}

// NewPorts will produces a new *Ports struct
func NewPorts() *Ports {
	return &Ports{specs: []string{}, publishall: true}
}

// Publish is intended to override an internal port with a specific external
// port.
//    ports := NewPorts()
//    ports.Publish(80, 8080) // Will expose the internal port 80 as 8080
func (ports *Ports) Publish(internal int, external int) {
	ports.specs = append(
		ports.specs, fmt.Sprintf("%d:%d",  internal, external))
}

// PublishAll is used to toggle the value for HostConfig.PublishAllPorts.
func (ports *Ports) PublishAll(enabled bool) {
	ports.publishall = enabled
}

// HostConfig converts the ports provided to Publish() and produces a HostConfig
// struct that can be used when creating the container.
func (ports *Ports) HostConfig() (*container.HostConfig, error) {
	config := &container.HostConfig{PublishAllPorts: ports.publishall}

	if len(ports.specs) > 0 {
		_, bindings, err := nat.ParsePortSpecs(ports.specs)
		if err != nil {
			return nil, err
		}

		config.PortBindings = bindings
	}
	return config, nil
}


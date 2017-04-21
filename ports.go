package dockertest

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// Ports is when to convey port exposures to RunContainer()
type Ports struct {
	specs      []string
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
		ports.specs, fmt.Sprintf("%d:%d", internal, external))
}

// PublishAll is used to toggle the value for HostConfig.PublishAllPorts.
func (ports *Ports) PublishAll(enabled bool) {
	ports.publishall = enabled
}

// HostConfig converts the struct to a *c.HostConfig which can be used
// to run containers.
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

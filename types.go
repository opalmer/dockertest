package dockertest

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
)

// types.go provides a small set of basic types that are used and returned
// by dockertest.


// Labels are used when creating a container to identify containers that are
// run by us.
type Labels struct {
	labels map[string]string
}

// NewLabels produces a *Labels
func NewLables() *Labels {
	return Labels{labels: map[string]string{}}
}

// Add is used to add a new label
func (labels *Labels) Add(key string, value string) {
	labels.labels[key] = value
}

// Remove is used to remove a label
func (labels *Labels) Remove(key string) {
	delete(labels.labels, key)
}

// Ports is when to convey port exposures to RunContainer()
type Ports struct {
	specs []string
	publishall bool
}

// NewPorts will produces a new *Ports struct.
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

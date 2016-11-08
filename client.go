package dockertest

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

)

// client.go provides a high level client for interacting with Docker

// DockerClient provides a wrapper for the standard docker client
type DockerClient struct {
	Client *client.Client
}

// NewDockerClient produces a new *DockerClient that can be used to interact
// with Docker.
func NewDockerClient() (*DockerClient , error) {
	docker, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &DockerClient{Client: docker}, nil
}

// RunContainer will run a new container and return the results. By default
// all ports that are exposed by the container will be published to the host.
func (docker *DockerClient) RunContainer(image string, labels *Labels, ports *Ports) error {
	hostconfig, err := ports.HostConfig()
	if err != nil {
		return err
	}

	if labels == nil {
		labels := NewLables()
		labels.Add("creator", "dockertest")
	}

	created, err := client.Client.ContainerCreate(
		context.Background(), &container.Config{
			Image:  image,
			Labels: labels,
		},
		hostconfig, &network.NetworkingConfig{}, "")
	return nil
}
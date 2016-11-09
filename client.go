package dockertest

import (
	"context"
	"errors"

	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// client.go provides a high level client for interacting with Docker

var (
	// ErrContainerNotFound is returned by GetContainer if we were
	// unable to find the requested container.
	ErrContainerNotFound = errors.New(
		"Expected to find exactly one container for the given query.")
)

// DockerClient provides a wrapper for the standard docker client
type DockerClient struct {
	Client *client.Client
	log    *log.Entry
}

// NewDockerClient produces a new *DockerClient that can be used to interact
// with Docker.
func NewDockerClient() (*DockerClient, error) {
	docker, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &DockerClient{Client: docker, log: log.WithField("phase", "client")}, nil
}

// GetContainer retrieves a single container by id.
func (docker *DockerClient) GetContainer(id string) (*Container, error) {
	args := filters.NewArgs()
	args.Add("id", id)
	options := types.ContainerListOptions{Filters: args}
	containers, err := client.Client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	if len(containers) != 1 {
		return nil, ErrContainerNotFound
	}

	return NewContainer(containers[0])
}

// Containers is used to return a list of filtered containers matching the given
// image and label. Note, this will only return running containers.
func (docker *DockerClient) Containers(image string, label string) ([]*Container, error) {
	args := filters.NewArgs()
	args.Add("ancestor", image)
	args.Add("label", fmt.Sprintf("%s=1", label))
	args.Add("label", "gerrittest=1")
	args.Add("status", "running")

	output := []*Container{}

	options := types.ContainerListOptions{Filters: args}
	containers, err := docker.Client.ContainerList(context.Background(), options)
	if err != nil {
		return output, err
	}

	for _, entry := range containers {

	}

}

// RunContainer will run a new container and return the results. By default
// all ports that are exposed by the container will be published to the host.
func (docker *DockerClient) RunContainer(image string, label string, ports *Ports) (*Container, error) {
	hostconfig, err := ports.HostConfig()
	if err != nil {
		return err
	}

	labels := map[string]string{}
	labels["dockertest"] = "1"
	labels[label] = "1"

	created, err := client.Client.ContainerCreate(
		context.Background(), &container.Config{
			Image: image, Labels: labels},
		hostconfig, &network.NetworkingConfig{}, "")

	err = docker.Client.ContainerStart(
		context.Background(), created.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	for _, warning := range created.Warnings {
		docker.log.Warn(warning)
	}

	return docker.GetContainer(created.ID)
}

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
	"io/ioutil"
	"io"
)

// client.go provides a high level client for interacting with Docker

var (
	// ErrContainerNotFound is returned by GetContainer if we were
	// unable to find the requested container.
	ErrContainerNotFound = errors.New(
		"Expected to find exactly one container for the given query.")
)

// DockerClient provides a wrapper for the standard dc client
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
	return &DockerClient{
		Client: docker, log: log.WithField("phase", "client")}, nil
}

// Container retrieves a single container by id and returns a *Container
// struct.
func (dc *DockerClient) Container(id string) (*Container, error) {
	args := filters.NewArgs()
	args.Add("id", id)
	options := types.ContainerListOptions{Filters: args}
	containers, err := dc.Client.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	if len(containers) != 1 {
		return nil, ErrContainerNotFound
	}

	return NewContainer(containers[0]), nil
}

// Containers is used to return a list of filtered containers matching the given
// image and label. The following criteria are used to filter the results from
// Docker.
//    ancestor=<image>
//    label=<label>=1
//    label=dockertest=1
//    status=running
func (dc *DockerClient) Containers(image string, label string) ([]*Container, error) {
	args := filters.NewArgs()
	args.Add("ancestor", image)
	args.Add("label", fmt.Sprintf("%s=1", label))
	args.Add("label", "dockertest=1")
	args.Add("status", "running")

	output := []*Container{}

	options := types.ContainerListOptions{Filters: args}
	containers, err := dc.Client.ContainerList(context.Background(), options)
	if err != nil {
		return output, err
	}

	for _, entry := range containers {
		output = append(output, NewContainer(entry))
	}
	return output, nil
}

// RunContainer will run a new container and return the results. By default
// all ports that are exposed by the container will be published to the host
// randomly. The published ports will be accessible using functions on the
// struct:
//    client, err := NewDockerClient()
//    container := client.RunContainer("testimage", "testing", nil)
//    port, err := container.Port(80)
//    port.External
func (dc *DockerClient) RunContainer(image string, label string, ports *Ports) (*Container, error) {
	logger := dc.log.WithFields(log.Fields{
		"image": image,
		"label": label,
	})

	if ports == nil {
		ports = NewPorts()
	}

	hostconfig, err := ports.HostConfig()
	if err != nil {
		return nil, err
	}

	labels := map[string]string{}
	labels["dockertest"] = "1"
	if label != "" {
		labels[label] = "1"
	}

	var created container.ContainerCreateCreatedBody
creation:
	for {
		logger = logger.WithField("action", "create")
		created, err = dc.Client.ContainerCreate(
			context.Background(),
			&container.Config{
				Image: image,
				Labels: labels,
			},
			hostconfig, &network.NetworkingConfig{}, "")
		switch {
		case client.IsErrNotFound(err):
			logger = logger.WithFields(log.Fields{
				"action": "pull-image",
			})
			logger.Info()
			reader, err := dc.Client.ImagePull(context.Background(), image, types.ImagePullOptions{})
			if err != nil {
				logger.WithError(err).Error()
				return nil, err
			}
			io.Copy(ioutil.Discard, reader)
		case err != nil:
			logger.Info("hedrerer")
			logger.WithError(err).Error()
			return nil, err
		case err == nil:
			break creation
		}

	}

	logger = logger.WithFields(log.Fields{
		"action": "start",
		"id": created.ID,
	})

	logger.Info()
	err = dc.Client.ContainerStart(
		context.Background(), created.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	for _, warning := range created.Warnings {
		logger.Warn(warning)
	}

	return dc.Container(created.ID)
}

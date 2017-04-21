package dockertest

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

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
	// unable to find the requested c.
	ErrContainerNotFound = errors.New(
		"Expected to find exactly one c for the given query.")
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

// ContainerInfo retrieves a single c by id and returns a *ContainerInfo
// struct.
func (dc *DockerClient) Container(ctx context.Context, id string) (*ContainerInfo, error) {
	args := filters.NewArgs()
	args.Add("id", id)
	options := types.ContainerListOptions{Filters: args}
	containers, err := dc.Client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, ErrContainerNotFound
	}

	inspection, err := dc.Client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return NewContainerInfo(containers[0], inspection), nil
}

func (dc *DockerClient) filter(ctx context.Context, cancel context.CancelFunc, input *ClientInput, containers chan *ContainerInfo) {
	options := types.ContainerListOptions{
		All:     input.All,
		Since:   input.Since,
		Before:  input.Before,
		Filters: input.FilterArgs(),
	}

	found, err := dc.Client.ContainerList(ctx, options)
	if err != nil {
		return
	}

	for _, entry := range found {
		//entry.ID
	}
}

// FilterContainers will iterate over all containers using the provided
// input and emit the results to a channel.
func (dc *DockerClient) FilterContainers(ctx context.Context, input *ClientInput) chan *ContainerInfo {
	ctx, cancel := context.WithCancel(ctx)
	containers := make(chan *ContainerInfo, 1)
	go dc.filter(ctx, cancel, input, containers)
	return containers, errs
}

// RunContainer will run a new c and return the results. By default
// all ports that are exposed by the c will be published to the host
// randomly. The published ports will be accessible using functions on the
// struct:
//    client, err := NewDockerClient()
//    c := client.RunContainer("testimage", "testing", nil)
//    port, err := c.Port(80)
//    port.External
func (dc *DockerClient) RunContainer(ctx context.Context, input *ClientInput) (*ContainerInfo, error) {
	logger := dc.log.WithFields(log.Fields{
		"image": input.Image,
	})

	hostconfig, err := input.Ports.HostConfig()
	if err != nil {
		return nil, err
	}

	var created container.ContainerCreateCreatedBody
creation:
	for {
		logger = logger.WithField("action", "create")
		created, err = dc.Client.ContainerCreate(
			ctx,
			input.ContainerConfig(),
			hostconfig, &network.NetworkingConfig{}, "")
		switch {
		case client.IsErrNotFound(err):
			logger = logger.WithFields(log.Fields{
				"action": "pull-image",
			})
			logger.Info()
			reader, err := dc.Client.ImagePull(context.Background(), input.Image, types.ImagePullOptions{})
			if err != nil {
				logger.WithError(err).Error()
				return nil, err
			}
			io.Copy(ioutil.Discard, reader)
		case err != nil:
			logger.WithError(err).Error()
			return nil, err
		case err == nil:
			break creation
		}

	}

	logger = logger.WithFields(log.Fields{
		"action": "start",
		"id":     created.ID,
	})

	logger.Info()
	err = dc.Client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	for _, warning := range created.Warnings {
		logger.Warn(warning)
	}

	return dc.Container(ctx, created.ID)
}

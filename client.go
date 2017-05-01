package dockertest

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/crewjam/errset"
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
	ErrContainerNotFound = errors.New("failed to locate the Container")
)

// DockerClient provides a wrapper for the standard dc client
type DockerClient struct {
	Client *client.Client
}

// NewClient produces a new *DockerClient that can be used to interact
// with Docker.
func NewClient() (*DockerClient, error) {
	docker, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &DockerClient{Client: docker}, nil
}

// ContainerInfo retrieves a single c by id and returns a *ContainerInfo
// struct.
func (dc *DockerClient) ContainerInfo(ctx context.Context, id string) (*ContainerInfo, error) {
	args := filters.NewArgs()
	args.Add("id", id)
	options := types.ContainerListOptions{Filters: args, All: true}
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

	return &ContainerInfo{
		Data:  containers[0],
		State: inspection.State, JSON: inspection,
		Warnings: []string{},
		client:   dc,
	}, nil
}

// ListContainers will return a list of *ContainerInfo structs based on the
// provided input.
func (dc *DockerClient) ListContainers(ctx context.Context, input *ClientInput) ([]*ContainerInfo, error) {
	options := types.ContainerListOptions{
		All:     input.All,
		Since:   input.Since,
		Before:  input.Before,
		Filters: input.FilterArgs(),
	}

	containers, err := dc.Client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}

	infos := make(chan *ContainerInfo)
	errs := make(chan error)

	for _, entry := range containers {
		go func(c types.Container) {
			info, err := dc.ContainerInfo(ctx, c.ID)
			if err != nil {
				errs <- err
				return
			}
			infos <- info
		}(entry)
	}

	results := []*ContainerInfo{}
	errout := errset.ErrSet{}
	for i := 0; i < len(containers); i++ {
		select {
		case err := <-errs:
			errout = append(errout, err)
		case info := <-infos:
			results = append(results, info)
		}
	}

	return results, errout.ReturnValue()
}

// RemoveContainer will delete the requested Container, force terminating
// it if necessary.
func (dc *DockerClient) RemoveContainer(ctx context.Context, id string) error {
	err := dc.Client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})

	// Docker's API does not expose their error structs and their
	// IsErrNotFound does not seem to work.
	if err != nil && strings.Contains(err.Error(), "No such container") {
		return nil
	}

	return err
}

// RunContainer will run a new c and return the results. By default
// all ports that are exposed by the c will be published to the host
// randomly. The published ports will be accessible using functions on the
// struct:
//    client, err := NewClient()
//    c := client.RunContainer("testimage", "testing", nil)
//    port, err := c.Port(80)
//    port.External
func (dc *DockerClient) RunContainer(ctx context.Context, input *ClientInput) (*ContainerInfo, error) {
	bindings, err := input.Ports.Bindings()
	if err != nil {
		return nil, err
	}

	hostconfig := &container.HostConfig{}
	hostconfig.PortBindings = bindings
	if err != nil {
		return nil, err
	}

	var created container.ContainerCreateCreatedBody
creation:
	for {
		created, err = dc.Client.ContainerCreate(
			ctx,
			input.ContainerConfig(),
			hostconfig, &network.NetworkingConfig{}, "")
		switch {
		case client.IsErrNotFound(err):
			reader, err := dc.Client.ImagePull(context.Background(), input.Image, types.ImagePullOptions{})
			if err != nil {
				return nil, err
			}
			io.Copy(ioutil.Discard, reader)
		case err != nil:
			return nil, err
		case err == nil:
			break creation
		}

	}

	err = dc.Client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	info, err := dc.ContainerInfo(ctx, created.ID)
	info.Warnings = created.Warnings
	return info, err
}

// Service will return a *Service struct that may be used to spin up
// a specific service. See the documentation present on the Service struct
// for more information.
func (dc *DockerClient) Service(input *ClientInput) *Service {
	return &Service{
		Input:  input,
		Client: dc,
	}
}

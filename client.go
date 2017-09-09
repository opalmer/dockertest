package dockertest

import (
	"context"
	"errors"
	"strings"

	"github.com/crewjam/errset"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var (
	// ErrContainerNotFound is returned by GetContainer if we were
	// unable to find the requested container.
	ErrContainerNotFound = errors.New("failed to locate the container")
)

// DockerClient provides a wrapper for the standard docker client. The intent
// is to wrap common operations so the internal of docker's own client are
// abstracted. Use NewClient() to construct and produce this struct.
type DockerClient struct {
	docker *client.Client
	ctx    context.Context
}

// ContainerInfo retrieves a single container by id and returns a
// *ContainerInfo struct.
func (d *DockerClient) ContainerInfo(id string) (*ContainerInfo, error) {
	args := filters.NewArgs()
	args.Add("id", id)

	options := types.ContainerListOptions{Filters: args, All: true}
	containers, err := d.docker.ContainerList(d.ctx, options)
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, ErrContainerNotFound
	}

	inspection, err := d.docker.ContainerInspect(d.ctx, id)
	if err != nil {
		return nil, err
	}

	return &ContainerInfo{
		Data:     containers[0],
		State:    inspection.State,
		JSON:     inspection,
		Warnings: []string{},
		client:   d,
	}, nil
}

// ListContainers will return a list of *ContainerInfo structs based on the
// provided input.
func (d *DockerClient) ListContainers(input *ClientInput) ([]*ContainerInfo, error) {
	options := types.ContainerListOptions{
		All:     input.All,
		Since:   input.Since,
		Before:  input.Before,
		Filters: input.FilterArgs(),
	}

	listed, err := d.docker.ContainerList(d.ctx, options)
	if err != nil {
		return nil, err
	}

	containers := make(chan *ContainerInfo)
	errs := make(chan error)

	for _, entry := range listed {
		go func(c types.Container) {
			info, err := d.ContainerInfo(c.ID)
			if err != nil {
				errs <- err
				return
			}
			containers <- info
		}(entry)
	}

	results := []*ContainerInfo{}
	errout := errset.ErrSet{}
	for i := 0; i < len(listed); i++ {
		select {
		case err := <-errs:
			errout = append(errout, err)
		case info := <-containers:
			results = append(results, info)
		}
	}

	return results, errout.ReturnValue()
}

// RemoveContainer will delete the requested Container, force terminating
// it if necessary.
func (d *DockerClient) RemoveContainer(id string) error {
	err := d.docker.ContainerRemove(d.ctx, id, types.ContainerRemoveOptions{Force: true})

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
func (d *DockerClient) RunContainer(input *ClientInput) (*ContainerInfo, error) {
	bindings, err := input.Ports.Bindings()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(d.ctx, DefaultServiceTimeout)
	defer cancel()

	if input.Timeout.Nanoseconds() > 0 {
		cancel()
		ctx, cancel = context.WithTimeout(d.ctx, input.Timeout)
		defer cancel()
	}

	for {
		created, err := d.docker.ContainerCreate(
			ctx,
			input.ContainerConfig(),
			&container.HostConfig{PortBindings: bindings}, &network.NetworkingConfig{}, "")
		if client.IsErrNotFound(err) {
			reader, err := d.docker.ImagePull(ctx, input.Image, types.ImagePullOptions{})
			if err != nil {
				return nil, err
			}
			reader.Close() // nolint: errcheck
			continue
		}
		if err != nil {
			return nil, err
		}
		if err := d.docker.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
			return nil, err
		}
		info, err := d.ContainerInfo(created.ID)
		info.Warnings = created.Warnings
		return info, err
	}
}

// Service will return a *Service struct that may be used to spin up
// a specific service. See the documentation present on the Service struct
// for more information.
func (d *DockerClient) Service(input *ClientInput) *Service {
	timeout := input.Timeout
	if timeout.Nanoseconds() == 0 {
		timeout = DefaultServiceTimeout
	}
	ctx, cancel := context.WithTimeout(d.ctx, timeout)
	go func() {
		defer cancel()
		<-ctx.Done()
	}()
	return &Service{Context: ctx, Input: input, Client: d}
}

// NewClient produces a *DockerClient struct.
func NewClient(ctx context.Context) (*DockerClient, error) {
	docker, err := client.NewEnvClient()
	return &DockerClient{docker: docker, ctx: ctx}, err
}

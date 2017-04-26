package dockertest

import (
	"context"
	"errors"
	"time"
)

const (
	// DefaultServiceTimeout is the default timeout that's applied
	// to all service operations.
	DefaultServiceTimeout = time.Minute * 3
)

// PingInput is used to provide inputs to a Ping function.
type PingInput struct {
	Service   *Service
	Container *ContainerInfo
}

// Ping is a function that's used to ping a service before returning from
// Service.Run. Any errors produced by ping will cause the associated
// container to be removed.
type Ping func(*PingInput) error

// Service is a struct used to run and manage a container for a specific
// service.
type Service struct {
	ctx       context.Context
	cancel    context.CancelFunc
	Image     string
	Timeout   time.Duration
	ping      Ping
	client    *DockerClient
	container *ContainerInfo
}

// Run will run the container.
func (s *Service) Run() error {
	ctx, cancel := context.WithTimeout(s.ctx, s.Timeout)
	defer cancel()

	runInput := NewClientInput(s.Image)

	info, err := s.client.RunContainer(ctx, runInput)
	if err != nil {
		return err
	}
	s.container = info
	input := &PingInput{
		Service:   s,
		Container: info,
	}
	if err := s.ping(input); err != nil {
		s.Terminate()
		return err
	}
	return nil
}

// Terminate terminates the container and returns.
func (s *Service) Terminate() error {
	if s.container == nil {
		return errors.New("Container not started")
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.Timeout)
	defer cancel()
	err := s.client.RemoveContainer(ctx, s.container.ID())
	s.cancel()
	return err
}

// Ping may be used to override the ping function. This has not effect if Run()
// has already been called.
func (s *Service) Ping(ping Ping) {
	s.ping = ping
}

// NewService produces a *Service struct.
func NewService(parent context.Context, image string) (*Service, error) {
	ctx, cancel := context.WithCancel(parent)
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Service{
		ctx:     ctx,
		cancel:  cancel,
		Image:   image,
		Timeout: DefaultServiceTimeout,
		client:  client,
	}, nil
}

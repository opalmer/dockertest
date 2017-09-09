package dockertest

import (
	"context"
	"errors"
	"time"

	"github.com/crewjam/errset"
)

const (
	// DefaultServiceTimeout is the default timeout that's applied
	// to all service operations.
	DefaultServiceTimeout = time.Minute * 3
)

var (
	// ErrInputNotProvided is returned by Service.Run if the Input field
	// is not provided.
	ErrInputNotProvided = errors.New("Input field not provided")

	// ErrContainerNotStarted is returned by Terminate() if the container
	// was never started.
	ErrContainerNotStarted = errors.New("Container not started")
)

// PingInput is used to provide inputs to a Ping function.
type PingInput struct {
	Service   *Service
	Container *ContainerInfo
}

// Ping is a function that's used to ping a service before returning from
// Service.Run. Any errors produced by ping will cause the associated
// Container to be removed.
type Ping func(*PingInput) error

// Service is a struct used to run and manage a Container for a specific
// service.
type Service struct {
	// Name is an optional name that may be used for tracking a service. This
	// field is not used by dockertest.
	Name string

	// Ping is a function that may be used to wait for the service
	// to come up before returning. If this function is specified
	// and it return an error Terminate() will be automatically
	// called. This function is called by Run() before returning.
	Ping Ping

	// Input is used to control the inputs to Run()
	Input *ClientInput

	// Timeout defines a duration that's used to prevent operations
	// related to docker from running forever. If this value is not
	// provided then DefaultServiceTimeout will be used.
	Timeout time.Duration

	// Client is the docker client.
	Client *DockerClient

	// Container will container information about the running container
	// once Run() has finished.
	Container *ContainerInfo
}

func (s *Service) timeout() time.Duration {
	if s.Timeout.Nanoseconds() != 0 {
		return s.Timeout
	}
	return DefaultServiceTimeout
}

// Run will run the Container.
func (s *Service) Run() error {
	if s.Input == nil {
		return ErrInputNotProvided
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout())
	defer cancel()

	info, err := s.Client.RunContainer(ctx, s.Input)
	if err != nil {
		return err
	}
	s.Container = info

	if s.Ping != nil {
		input := &PingInput{
			Service:   s,
			Container: info,
		}
		if err := s.Ping(input); err != nil {
			errs := errset.ErrSet{}
			errs = append(errs, err)
			errs = append(errs, s.Terminate())
			return errs.ReturnValue()
		}
	}

	return nil
}

// Terminate terminates the Container and returns.
func (s *Service) Terminate() error {
	if s.Container == nil {
		return ErrContainerNotStarted
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout())
	defer cancel()
	return s.Client.RemoveContainer(ctx, s.Container.ID())
}

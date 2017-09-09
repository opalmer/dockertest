package dockertest_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/opalmer/dockertest"
	. "gopkg.in/check.v1"
)

type ExamplesTest struct{}

var _ = Suite(&ExamplesTest{})

func (s *ExamplesTest) TestExampleNewClient(c *C) {
	ExampleNewClient()
}

func (s *ExamplesTest) TestExampleDockerClient_Service(c *C) {
	ExampleDockerClient_Service()
}

//
// NOTE:
//     If you modify the example functions below, please make sure to
//     update README.md too.
//

func ExampleNewClient() {
	client, err := dockertest.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Construct information about the container to start.
	input := dockertest.NewClientInput("nginx:mainline-alpine")
	input.Ports.Add(&dockertest.Port{
		Private:  80,
		Public:   dockertest.RandomPort,
		Protocol: dockertest.ProtocolTCP,
	})

	// Start the container
	container, err := client.RunContainer(context.Background(), input)
	if err != nil {
		log.Fatal(err)
	}

	// Extract information about the started container.
	port, err := container.Port(80)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(port.Public, port.Address)

	if err := client.RemoveContainer(context.Background(), container.ID()); err != nil {
		log.Fatal(err)
	}
}

func ExampleDockerClient_Service() {
	client, err := dockertest.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Construct information about the container to start.
	input := dockertest.NewClientInput("nginx:mainline-alpine")
	input.Ports.Add(&dockertest.Port{
		Private:  80,
		Public:   dockertest.RandomPort,
		Protocol: dockertest.ProtocolTCP,
	})

	// Construct the service and tell it how to handle waiting
	// for the container to start.
	service := client.Service(input)
	service.Ping = func(input *dockertest.PingInput) error {
		port, err := input.Container.Port(80)
		if err != nil {
			return err // Will cause Run() to call Terminate()
		}

		for {
			_, err := net.Dial(string(port.Protocol), fmt.Sprintf("%s:%d", port.Address, port.Public))
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			break
		}

		return nil
	}

	// Starts the container, runs Ping() and waits for it to return. If Ping()
	// fails the container will be terminated and Run() will return an error.
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}

	// Container has started, get information information
	// about the exposed port.
	port, err := service.Container.Port(80)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(port.Public, port.Address)

	if err := service.Terminate(); err != nil {
		log.Fatal(err)
	}
}

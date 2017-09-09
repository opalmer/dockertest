package dockertest

import (
	"context"
	"fmt"
	"log"

	. "gopkg.in/check.v1"
)

type ExamplesTest struct{}

var _ = Suite(&ExamplesTest{})

func (s *ExamplesTest) TestExampleCreateContainerWithPort(c *C) {
	ExampleCreateContainerWithPort()
}

func ExampleCreateContainerWithPort() {
	client, err := NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Construct information about the container to start.
	input := NewClientInput("nginx:mainline-alpine")
	input.Ports.Add(&Port{
		Private:  80,
		Public:   RandomPort,
		Protocol: ProtocolTCP,
	})

	// Start the container
	container, err := client.RunContainer(input)
	if err != nil {
		log.Fatal(err)
	}

	// Extract information about the started container.
	port, err := container.Port(80)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(port.Public, port.Address)
}

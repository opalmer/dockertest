# Docker Test

[![Build Status](https://travis-ci.org/opalmer/dockertest.svg?branch=master)](https://travis-ci.org/opalmer/dockertest)
[![codecov](https://codecov.io/gh/opalmer/dockertest/branch/master/graph/badge.svg)](https://codecov.io/gh/opalmer/dockertest)


This project provides a small set of wrappers around docker. It is intended
to be used to ease testing. Documentation is available via godoc: 
    https://godoc.org/github.com/opalmer/dockertest

# Examples

Create a container and retrieve an exposed port.

```go
import "github.com/opalmer/dockertest"

func main() {
	client, err := dockertest.NewClient()
	input := dockertest.NewClientInput("nginx")
	input.Ports.Add(&dockertest.Port{
		Private: 80,
		Public: dockertest.RandomPort,
		Protocol: dockertest.ProtocolTCP,
	})
	port, err := container.Port(80)
	fmt.Println(port.Public, port.Address)
}
```

Create a container using the `Service` struct.

```go
import (
	"net"
	"github.com/opalmer/dockertest"
)

func main() {
	client, _ := dockertest.NewClient()
	service := client.Service("nginx")
	service.Ping = func(input *dockertest.PingInput) error {
		port, err := input.Container.Port(80)
		if err != nil {
			return err // Will cause Run() to call Terminate()
		}

		for {
			_, err := net.Dial(string(port.Protocol), fmt.Sprintf("%s:%d", port.Address, port.Public))
			if err != nil {
				return nil
			}
		}
	}
	err := service.Run() // Will return when Ping() returns
	defer service.Terminate()
}
```

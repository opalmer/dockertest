# Docker Test

This project provides a small set of wrappers around docker. It is intended
to be used to ease testing.

# Example

```go
import "github.com/opalmer/dockertest"

func main() {
	client, _ := dockertest.NewDockerClient()
	container, _ = client.RunContainer("image", "label", nil)
	
	// Retrieve the public port for a port which the container 
	// exposes. Useful when spinning up containers running services 
	// you're testing.
	port, _ := container.Port(5555)
	port.PublicPort
}
```
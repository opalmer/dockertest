package dockertest

import (
	"github.com/docker/go-connections/nat"
	. "gopkg.in/check.v1"
)

type TestPorts struct{}

var _ = Suite(&TestPorts{})

func (s *TestPorts) TestPublish(c *C) {
	ports := NewPorts()
	ports.Publish(80, 8080)
	c.Assert(ports.specs[0], Equals, "80:8080")
}

func (s *TestPorts) TestPublishAll(c *C) {
	ports := NewPorts()
	c.Assert(ports.publishall, Equals, true)
	ports.PublishAll(false)
	c.Assert(ports.publishall, Equals, false)
}

func (s *TestPorts) TestHostConfig(c *C) {
	ports := NewPorts()
	ports.Publish(80, 8080)

	hostconfig, err := ports.HostConfig()
	c.Assert(err, IsNil)
	c.Assert(hostconfig.PublishAllPorts, Equals, true)

	c.Assert(
		hostconfig.PortBindings, DeepEquals,
		nat.PortMap{"8080/tcp": []nat.PortBinding{{HostIP: "", HostPort: "80"}}})
}

package dockertest

import (
	"io/ioutil"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-connections/nat"
	. "gopkg.in/check.v1"
)

type TestSuite struct{}

func TestDockerTest(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	TestingT(t)
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestPorts_Publish(c *C) {
	ports := NewPorts()
	ports.Publish(80, 8080)
	c.Assert(ports.specs[0], Equals, "80:8080")
}

func (s *TestSuite) TestPorts_PublishAll(c *C) {
	ports := NewPorts()
	c.Assert(ports.publishall, Equals, true)
	ports.PublishAll(false)
	c.Assert(ports.publishall, Equals, false)
}

func (s *TestSuite) TestPorts_HostConfig(c *C) {
	ports := NewPorts()
	ports.Publish(80, 8080)

	hostconfig, err := ports.HostConfig()
	c.Assert(err, IsNil)
	c.Assert(hostconfig.PublishAllPorts, Equals, true)

	c.Assert(
		hostconfig.PortBindings, DeepEquals,
		nat.PortMap{"8080/tcp": []nat.PortBinding{{HostIP: "", HostPort: "80"}}})
}

package dockertest

import (
	"github.com/docker/go-connections/nat"
	. "gopkg.in/check.v1"
)

type TestPorts struct{}

var _ = Suite(&TestPorts{})

func (s *TestPorts) TestNewPorts(c *C) {
	ports := NewPorts()
	c.Assert(ports, DeepEquals, &Ports{Specs: []*Port{}})
}

func (s *TestPorts) TestPortPort(c *C) {
	port := &Port{
		Protocol: ProtocolUDP,
		Private:  5555,
	}

	expected, err := nat.NewPort(string(ProtocolUDP), "5555")
	np, err := port.Port()
	c.Assert(err, IsNil)
	c.Assert(np, Equals, expected)
}

func (s *TestPorts) TestPortBinding(c *C) {
	port := &Port{
		Protocol: ProtocolUDP,
		Address:  "1.2.3.4",
		Public:   1234,
	}
	c.Assert(port.Binding(), DeepEquals, nat.PortBinding{HostIP: "1.2.3.4", HostPort: "1234"})
	port = &Port{
		Protocol: ProtocolUDP,
		Public:   1234,
	}
	c.Assert(port.Binding(), DeepEquals, nat.PortBinding{HostIP: "0.0.0.0", HostPort: "1234"})
}

func (s *TestPorts) TestPortsAdd(c *C) {
	ports := NewPorts()
	port := &Port{
		Protocol: ProtocolUDP,
		Address:  "1.2.3.4",
		Public:   1234,
	}
	ports.Add(port)
	c.Assert(ports.Specs, DeepEquals, []*Port{port})
}

func (s *TestPorts) TestPortsBindings(c *C) {
	ports := NewPorts()
	port := &Port{
		Protocol: ProtocolUDP,
		Address:  "1.2.3.4",
		Public:   1234,
		Private:  4567,
	}
	ports.Add(port)
	bindings, err := ports.Bindings()
	c.Assert(err, IsNil)
	mapping := nat.PortMap{}
	mapping["4567/udp"] = []nat.PortBinding{port.Binding()}
	c.Assert(bindings, DeepEquals, mapping)
}

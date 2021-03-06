package dockertest

import (
	"errors"
	"fmt"
	"net"

	. "gopkg.in/check.v1"
)

type ServiceTest struct{}

var _ = Suite(&ServiceTest{})

func (*ServiceTest) TestNoInput(c *C) {
	s := &Service{}
	c.Assert(s.Run(), ErrorMatches, ErrInputNotProvided.Error())
}

func (*ServiceTest) TestRunWithPing(c *C) {
	dc, err := NewClient()
	c.Assert(err, IsNil)
	defer dc.docker.Close() // nolint: errcheck

	input := NewClientInput(testImage)
	input.Ports.Add(&Port{
		Private:  80,
		Public:   RandomPort,
		Protocol: ProtocolTCP,
	})
	svc := dc.Service(input)
	svc.Ping = func(input *PingInput) error {
		port, err := input.Container.Port(80)
		c.Assert(err, IsNil)
		for {
			con, err := net.Dial(string(port.Protocol), fmt.Sprintf("%s:%d", port.Address, port.Public))
			if err != nil {
				continue
			}
			defer con.Close() // nolint: errcheck
			return nil
		}
	}
	c.Assert(svc.Run(), IsNil)
	c.Assert(svc.Terminate(), IsNil)
}

func (*ServiceTest) TestErrorOnPingCallsTerminate(c *C) {
	dc, err := NewClient()
	c.Assert(err, IsNil)
	defer dc.docker.Close() // nolint: errcheck

	input := NewClientInput(testImage)
	svc := dc.Service(input)
	svc.Ping = func(input *PingInput) error {
		return errors.New("some error")
	}
	c.Assert(svc.Run(), ErrorMatches, "some error")
	c.Assert(svc.Terminate(), IsNil)
}

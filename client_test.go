package dockertest

import (
	"context"
	"fmt"
	"os"
	"time"

	. "gopkg.in/check.v1"
)

type ClientTest struct{}

var _ = Suite(&ClientTest{})

func (s *ClientTest) TestNewClient(c *C) {
	_, err := NewClient()
	c.Assert(err, IsNil)

	dockerhost := os.Getenv("DOCKER_HOST")
	if dockerhost != "" {
		defer os.Setenv("DOCKER_HOST", dockerhost)
	} else {
		defer os.Unsetenv("DOCKER_HOST")
	}

	os.Setenv("DOCKER_HOST", "/////")
	_, err = NewClient()
	c.Assert(err, ErrorMatches, "unable to parse docker host `/////`")
}

func (s *ClientTest) TestRunAndRemoveContainer(c *C) {
	dc, err := NewClient()
	c.Assert(err, IsNil)
	input := NewClientInput("rabbitmq:3")

	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	defer dc.RemoveContainer(context.Background(), info.Data.ID)
}

func (s *ClientTest) TestRemoveContainer(c *C) {
	dc, err := NewClient()
	c.Assert(err, IsNil)
	input := NewClientInput("rabbitmq:3")
	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
}

func (s *ClientTest) TestListContainers(c *C) {
	dc, err := NewClient()
	c.Assert(err, IsNil)

	label := fmt.Sprintf("%d", time.Now().Nanosecond())
	infos := map[string]*ContainerInfo{}
	for i := 0; i < 4; i++ {
		input := NewClientInput("rabbitmq:3")
		input.SetLabel("time", label)
		info, err := dc.RunContainer(context.Background(), input)
		c.Assert(err, IsNil)
		infos[info.Data.ID] = info
		defer dc.RemoveContainer(context.Background(), info.Data.ID)
	}

	input := NewClientInput("rabbitmq:3")
	input.SetLabel("time", label)

	containers, err := dc.ListContainers(context.Background(), input)
	c.Assert(err, IsNil)
	for _, entry := range containers {
		_, ok := infos[entry.Data.ID]
		c.Assert(ok, Equals, true)

	}
}

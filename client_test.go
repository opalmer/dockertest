package dockertest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type ClientTest struct{}

var _ = Suite(&ClientTest{})

func (s *ClientTest) TestNewClient(c *C) {
	dc, err := NewClient(context.Background())
	defer dc.docker.Close() // nolint: errcheck
	c.Assert(err, IsNil)

	dockerhost := os.Getenv("DOCKER_HOST")
	if dockerhost != "" {
		defer os.Setenv("DOCKER_HOST", dockerhost) // nolint: errcheck
	} else {
		defer os.Unsetenv("DOCKER_HOST") // nolint: errcheck
	}

	c.Assert(os.Setenv("DOCKER_HOST", "/////"), IsNil)
	_, err = NewClient(context.Background())
	c.Assert(err, ErrorMatches, "unable to parse docker host `/////`")
}

func (s *ClientTest) TestRunAndRemoveContainer(c *C) {
	dc, err := NewClient(context.Background())
	defer dc.docker.Close() // nolint: errcheck
	c.Assert(err, IsNil)
	input := NewClientInput(testImage)

	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	c.Assert(info.Refresh(), IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
}

func (s *ClientTest) TestRunContainerAttemptsToRetrieveImage(c *C) {
	dc, err := NewClient(context.Background())
	c.Assert(err, IsNil)

	// Some random image so it will force the client to try to pull the
	// image down.
	input := NewClientInput("abcdefgzyn")
	_, err = dc.RunContainer(context.Background(), input)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "does not exist"), Equals, true)
}

func (s *ClientTest) TestRemoveContainer(c *C) {
	dc, err := NewClient(context.Background())
	defer dc.docker.Close() // nolint: errcheck
	c.Assert(err, IsNil)
	input := NewClientInput(testImage)
	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
}

func (s *ClientTest) TestListContainers(c *C) {
	dc, err := NewClient(context.Background())
	c.Assert(err, IsNil)
	defer dc.docker.Close() // nolint: errcheck

	label := fmt.Sprintf("%d", time.Now().Nanosecond())
	infos := map[string]*ContainerInfo{}
	for i := 0; i < 4; i++ {
		input := NewClientInput(testImage)
		input.SetLabel("time", label)
		info, err := dc.RunContainer(context.Background(), input)
		c.Assert(err, IsNil)
		infos[info.Data.ID] = info
	}

	input := NewClientInput(testImage)
	input.SetLabel("time", label)

	containers, err := dc.ListContainers(context.Background(), input)
	c.Assert(err, IsNil)
	for _, entry := range containers {
		_, ok := infos[entry.Data.ID]
		c.Assert(ok, Equals, true)
	}

	for key := range infos {
		c.Assert(dc.RemoveContainer(context.Background(), key), IsNil)
	}
}

func (s *ClientTest) TestService(c *C) {
	dc, err := NewClient(context.Background())
	c.Assert(err, IsNil)
	defer dc.docker.Close() // nolint: errcheck
	input := NewClientInput(testImage)
	svc := dc.Service(input)
	c.Assert(svc.Input, DeepEquals, input)
}

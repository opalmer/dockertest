package dockertest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type ClientTest struct {
	cleanups []func() error
}

var _ = Suite(&ClientTest{})

func (s *ClientTest) TearDownSuite(c *C) {
	for _, function := range s.cleanups {
		c.Assert(function(), IsNil)
	}
}

func (s *ClientTest) addCleanup(f func() error) {
	s.cleanups = append(s.cleanups, f)
}

func (s *ClientTest) newClient(c *C) *DockerClient {
	dc, err := NewClient()
	c.Assert(err, IsNil)
	return dc
}

func (s *ClientTest) TestNewClient(c *C) {
	dc := s.newClient(c)
	dc, err := NewClient()
	s.addCleanup(dc.docker.Close)
	c.Assert(err, IsNil)

	dockerhost := os.Getenv("DOCKER_HOST")
	if dockerhost != "" {
		defer os.Setenv("DOCKER_HOST", dockerhost) // nolint: errcheck
	} else {
		defer os.Unsetenv("DOCKER_HOST") // nolint: errcheck
	}

	c.Assert(os.Setenv("DOCKER_HOST", "/////"), IsNil)
	_, err = NewClient()
	c.Assert(err, ErrorMatches, "unable to parse docker host `/////`")
}

func (s *ClientTest) TestRunAndRemoveContainer(c *C) {
	dc := s.newClient(c)
	input := NewClientInput(testImage)

	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	c.Assert(info.Refresh(), IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
}

func (s *ClientTest) TestRunContainerAttemptsToRetrieveImage(c *C) {
	dc := s.newClient(c)

	// Some random image so it will force the client to try to pull the
	// image down.
	input := NewClientInput("988881adc9fc3655077dc2d4d757d480b5ea0e11")
	_, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "does not exist"), Equals, true)
}

func (s *ClientTest) TestRemoveContainer(c *C) {
	dc := s.newClient(c)
	input := NewClientInput(testImage)
	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
}

func (s *ClientTest) TestContainerInfo(c *C) {
	dc := s.newClient(c)
	_, err := dc.ContainerInfo(context.Background(), "foobar")
	c.Assert(err, ErrorMatches, ErrContainerNotFound.Error())
}

func (s *ClientTest) TestListContainers(c *C) {
	dc := s.newClient(c)

	label := fmt.Sprintf("%d", time.Now().Nanosecond())
	infos := map[string]*ContainerInfo{}
	for i := 0; i < 4; i++ {
		input := NewClientInput(testImage)
		info, err := dc.RunContainer(context.Background(), input)
		infos[info.Data.ID] = info
		c.Assert(err, IsNil)
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

func (s *ClientTest) Test_getContainerInfo_ErrContainerNotFound(c *C) {
	dc := s.newClient(c)
	errs := make(chan error, 1)
	containers := make(chan *ContainerInfo)
	dc.getContainerInfo(context.Background(), "hello", containers, errs)
	c.Assert(<-errs, ErrorMatches, ErrContainerNotFound.Error())
}

func (s *ClientTest) Test_getContainerInfo(c *C) {
	dc := s.newClient(c)
	errs := make(chan error)
	containers := make(chan *ContainerInfo, 1)
	input := NewClientInput(testImage)
	info, err := dc.RunContainer(context.Background(), input)
	c.Assert(err, IsNil)
	dc.getContainerInfo(context.Background(), info.ID(), containers, errs)
	c.Assert(dc.RemoveContainer(context.Background(), info.ID()), IsNil)
	container := <-containers
	c.Assert(container.ID(), Equals, info.ID())
}

func (s *ClientTest) TestListContainersContainersRemoved(c *C) {
	dc := s.newClient(c)

	label := fmt.Sprintf("%d", time.Now().Nanosecond())
	infos := map[string]*ContainerInfo{}
	for i := 0; i < 4; i++ {
		input := NewClientInput(testImage)
		info, err := dc.RunContainer(context.Background(), input)
		c.Assert(err, IsNil)
		infos[info.Data.ID] = info
		c.Assert(dc.RemoveContainer(context.Background(), info.Data.ID), IsNil)
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
	dc := s.newClient(c)
	input := NewClientInput(testImage)
	svc := dc.Service(input)
	c.Assert(svc.Input, DeepEquals, input)
}

package dockertest

import (
	"os"
	"time"

	"github.com/docker/docker/api/types"
	. "gopkg.in/check.v1"
)

type ContainerInfoTest struct{}

var _ = Suite(&ContainerInfoTest{})

func (s *ContainerInfoTest) TestString(c *C) {
	info := &ContainerInfo{
		Data: types.Container{
			Image:  "image",
			ID:     "id",
			Status: "status",
		},
	}
	c.Assert(info.String(), Equals, "{image:image, id:id, status:status}")
}

func (s *ContainerInfoTest) TestHasLabel(c *C) {
	info := &ContainerInfo{
		Data: types.Container{
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	c.Assert(info.HasLabel("foo", "bar"), Equals, true)
	c.Assert(info.HasLabel("foo", ""), Equals, false)
}

func (s *ContainerInfoTest) TestPort(c *C) {
	info := &ContainerInfo{
		Data: types.Container{
			Ports: []types.Port{{
				PrivatePort: 50000,
				PublicPort:  2,
				Type:        "udp",
			}},
		},
	}
	port, err := info.Port(50000)
	c.Assert(err, IsNil)
	c.Assert(port.Public, Equals, uint16(2))
	_, err = info.Port(12)
	c.Assert(err, ErrorMatches, ErrPortNotFound.Error())
}

func (s *ContainerInfoTest) TestID(c *C) {
	info := &ContainerInfo{
		Data: types.Container{
			ID: "foobar",
		},
	}
	c.Assert(info.ID(), Equals, "foobar")
}

func (s *ContainerInfoTest) TestStarted(c *C) {
	info := &ContainerInfo{
		State: &types.ContainerState{
			StartedAt: timeNotSet,
		},
	}
	_, err := info.Started()
	c.Assert(err, ErrorMatches, ErrContainerNotRunning.Error())

	now := time.Now()
	info = &ContainerInfo{
		State: &types.ContainerState{
			StartedAt: now.Format(time.RFC3339Nano),
		},
	}
	value, err := info.Started()
	c.Assert(err, IsNil)
	c.Assert(value.UnixNano(), Equals, now.UnixNano())
}

func (s *ContainerInfoTest) TestFinished(c *C) {
	info := &ContainerInfo{
		State: &types.ContainerState{
			FinishedAt: timeNotSet,
		},
	}
	_, err := info.Finished()
	c.Assert(err, ErrorMatches, ErrContainerStillRunning.Error())

	now := time.Now()
	info = &ContainerInfo{
		State: &types.ContainerState{
			FinishedAt: now.Format(time.RFC3339Nano),
		},
	}
	value, err := info.Finished()
	c.Assert(err, IsNil)
	c.Assert(value.UnixNano(), Equals, now.UnixNano())
}

func (s *ContainerInfoTest) TestElapsed(c *C) {
	toValue := func(t time.Time) string {
		return t.Format(time.RFC3339Nano)
	}
	expectations := map[*ContainerInfo]time.Duration{
		{
			State: &types.ContainerState{
				StartedAt:  timeNotSet,
				FinishedAt: timeNotSet,
			},
		}: time.Second * 0,
		{
			State: &types.ContainerState{
				StartedAt:  toValue(time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC)),
				FinishedAt: toValue(time.Date(2017, time.January, 1, 1, 0, 0, 0, time.UTC)),
			},
		}: time.Hour * 1,
	}

	for info, expected := range expectations {
		value, err := info.Elapsed()
		c.Assert(err, IsNil)
		c.Assert(value.Nanoseconds(), Equals, expected.Nanoseconds())
	}
}

func (s *ContainerInfoTest) TestAddressDockerURL(c *C) {
	current, set := os.LookupEnv("DOCKER_URL")
	if set {
		defer os.Setenv("DOCKER_URL", current) // nolint: errcheck
	} else {
		defer os.Unsetenv("DOCKER_URL") // nolint: errcheck
	}
	c.Assert(os.Setenv("DOCKER_URL", "tcp://1.2.3.4:80/"), IsNil)
	info := &ContainerInfo{}
	value, err := info.address("0.0.0.0")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "1.2.3.4")
}

func (s *ContainerInfoTest) TestAddressCannotParseDockerURL(c *C) {
	current, set := os.LookupEnv("DOCKER_URL")
	if set {
		defer os.Setenv("DOCKER_URL", current) // nolint: errcheck
	} else {
		defer os.Unsetenv("DOCKER_URL") // nolint: errcheck
	}
	c.Assert(os.Setenv("DOCKER_URL", "1.2.3.4:80/"), IsNil)
	info := &ContainerInfo{}
	value, err := info.address("0.0.0.0")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `parse "1.2.3.4:80/": first path segment in URL cannot contain colon`)
	c.Assert(value, Equals, "")
}

func (s *ContainerInfoTest) TestAddressIPAlreadySet(c *C) {
	current, set := os.LookupEnv("DOCKER_URL")
	if set {
		defer os.Setenv("DOCKER_URL", current) // nolint: errcheck
	}
	c.Assert(os.Unsetenv("DOCKER_URL"), IsNil)
	info := &ContainerInfo{}
	value, err := info.address("1.2.3.4")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "1.2.3.4")
}

func (s *ContainerInfoTest) TestAddressDefault(c *C) {
	current, set := os.LookupEnv("DOCKER_URL")
	if set {
		defer os.Setenv("DOCKER_URL", current) // nolint: errcheck
	}
	c.Assert(os.Unsetenv("DOCKER_URL"), IsNil)
	info := &ContainerInfo{}
	value, err := info.address("0.0.0.0")
	c.Assert(err, IsNil)
	c.Assert(value, Equals, "127.0.0.1")
}

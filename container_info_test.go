package dockertest

import (
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
			}},
		},
	}
	port, err := info.Port(50000)
	c.Assert(err, IsNil)
	c.Assert(port.PublicPort, Equals, uint16(2))
	_, err = info.Port(12)
	c.Assert(err, ErrorMatches, ErrPortNotFound.Error())
}

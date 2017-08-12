package dockertest

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	. "gopkg.in/check.v1"
)

type ClientInputsTest struct{}

var _ = Suite(&ClientInputsTest{})

func (s *ClientInputsTest) TestAddEnvironmentVar(c *C) {
	input := NewClientInput("test")
	input.AddEnvironmentVar("foo", "bar")
	c.Assert(input.Environment, DeepEquals, []string{"foo=bar"})
}

func (s *ClientInputsTest) TestSetLabel(c *C) {
	input := NewClientInput("test")
	input.SetLabel("foo", "bar")
	value, ok := input.Labels["foo"]
	c.Assert(ok, Equals, true)
	c.Assert(value, Equals, "bar")
}

func (s *ClientInputsTest) TestRemoveLabel(c *C) {
	input := NewClientInput("test")
	input.SetLabel("foo", "bar")
	input.RemoveLabel("foo")
	_, ok := input.Labels["foo"]
	c.Assert(ok, Equals, false)
}

func (s *ClientInputsTest) TestContainerConfig(c *C) {
	input := NewClientInput("test")
	input.SetLabel("foo", "bar")
	c.Assert(input.ContainerConfig(), DeepEquals, &container.Config{
		Image: "test",
		Labels: map[string]string{
			"foo":        "bar",
			"dockertest": "1",
		},
	})
}

func (s *ClientInputsTest) TestFilterArgs(c *C) {
	input := NewClientInput("test")
	input.SetLabel("foo", "bar")
	input.Status = "running"

	expected := filters.NewArgs()
	expected.Add("ancestor", "test")
	expected.Add("label", "dockertest=1")
	expected.Add("label", "foo=bar")
	expected.Add("status", "running")

	c.Assert(input.FilterArgs(), DeepEquals, expected)
}

func (s *ClientInputsTest) TestNewClientInput(c *C) {
	input := NewClientInput("test")
	c.Assert(input, DeepEquals, &ClientInput{
		Image: "test",
		Ports: NewPorts(),
		Labels: map[string]string{
			"dockertest": "1",
		},
	})
}

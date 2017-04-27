package dockertest

import (
	"testing"

	"gopkg.in/check.v1"
)

const testImage = "nginx:mainline-alpine"

func Test(t *testing.T) {
	check.TestingT(t)
}

PACKAGES = $(shell go list .)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./...)

DEPENDENCIES = \
    gopkg.in/check.v1 \
    golang.org/x/tools/cmd/goimports \
    github.com/golang/lint/golint \
    github.com/crewjam/errset \
    github.com/docker/docker/api/types \
    github.com/docker/docker/api/types/container \
    github.com/docker/docker/api/types/filters \
    github.com/docker/docker/api/types/network \
    github.com/docker/docker/client \
    github.com/docker/docker/api/types/container \
    github.com/docker/go-connections/nat

check: deps vet lint test

deps:
	go get $(DEPENDENCIES)
	rm -rf $(GOPATH)/src/github.com/docker/docker/vendor

lint: deps
	golint -set_exit_status $(PACKAGES)

fmt: deps
	go fmt $(PACKAGES)
	goimports -w $(PACKAGE_DIRS)

vet:
	go vet $(PACKAGES)

test: deps
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)

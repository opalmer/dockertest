PACKAGES = $(shell go list . | grep -v /vendor/)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./... | grep -v /vendor/)
SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/kardianos/govendor \
    github.com/golang/lint/golint

check: deps vet lint test

deps:
	go get $(EXTRA_DEPENDENCIES)
	govendor sync
	rm -rf $(GOPATH)/src/github.com/docker/docker/vendor
	rm -rf vendor/github.com/docker/docker/vendor

lint: deps
	golint -set_exit_status $(PACKAGES)

fmt: deps
	go fmt $(PACKAGES)
	goimports -w $(SOURCES)

vet:
	go vet $(PACKAGES)

test: deps
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)

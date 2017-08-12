PACKAGES = $(shell go list . | grep -v /vendor/)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./... | grep -v /vendor/)
SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/kardianos/govendor \
    github.com/golang/lint/golint \
    github.com/golang/dep/cmd/dep

check: deps vet lint test

deps:
	go get -u $(EXTRA_DEPENDENCIES)
	dep ensure

lint:
	golint -set_exit_status $(PACKAGES)

fmt:
	go fmt $(PACKAGES)
	goimports -w $(SOURCES)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)

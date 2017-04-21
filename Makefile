PACKAGES = $(shell go list . | grep -v /vendor/)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./... | egrep -v /vendor/)

check: fmt vet lint test

deps:
	[ -f $(GOPATH)/bin/golint ] || go get github.com/golang/lint/golint
	[ -f $(GOPATH)/bin/goimports ] || go get golang.org/x/tools/cmd/goimports

lint: deps
	golint -set_exit_status $(PACKAGES)

fmt: deps
	go fmt $(PACKAGES)
	goimports -w $(PACKAGE_DIRS)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)


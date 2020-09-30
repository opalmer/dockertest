PACKAGES = $(shell go list ./... )
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./...)
SOURCES = $(shell for f in $(PACKAGES); do ls $(shell go env GOPATH)/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/golang/lint/golint \
    github.com/tools/godep \
    github.com/alecthomas/gometalinter

check: deps vet lint test

deps:
	go get $(EXTRA_DEPENDENCIES)
	gometalinter --install > /dev/null

lint:
	gometalinter --vendor --disable-all --enable=deadcode --enable=errcheck --enable=goimports \
	--enable=gocyclo --enable=golint --enable=gosimple --enable=misspell \
	--enable=unconvert --enable=unused --enable=varcheck --enable=interfacer \
	./...

fmt:
	go fmt ./...
	goimports -w $(SOURCES)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)

PACKAGES = $(shell go list ./... )
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./...)
SOURCES = $(shell for f in $(PACKAGES); do ls $(shell go env GOPATH)/src/$$f/*.go; done)

check: vet lint test

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w $(SOURCES)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)
	go test -short -v ./... -test.failfast -test.count 10  # Flake

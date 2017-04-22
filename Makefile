PACKAGES = $(shell go list .)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./...)

check: vet lint test

lint:
	[ -f $(GOPATH)/bin/golint ] || go get github.com/golang/lint/golint
	golint -set_exit_status $(PACKAGES)

fmt:
	[ -f $(GOPATH)/bin/goimports ] || go get golang.org/x/tools/cmd/goimports
	go fmt $(PACKAGES)
	goimports -w $(PACKAGE_DIRS)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)


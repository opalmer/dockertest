PACKAGES = $(shell go list . | grep -v vendor)

check: fmt vet lint test

lint:
	golint || go get github.com/golang/lint/golint
	golint -set_exit_status $(PACKAGES)

fmt:
	go fmt $(PACKAGES)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)

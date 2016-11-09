PACKAGES = $(shell go list . | grep -v vendor)

check: fmt vet lint

lint:
	golint || go get github.com/golang/lint/golint
	golint -set_exit_status $(PACKAGES)

fmt:
	go fmt $(PACKAGES)

vet:
	go vet $(PACKAGES)
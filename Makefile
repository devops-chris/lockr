.PHONY: build install test lint clean release snapshot

BINARY_NAME=lockr
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

## install: Install to GOPATH/bin
install:
	go install $(LDFLAGS) .

## test: Run tests
test:
	go test -v ./...

## lint: Run linter
lint:
	golangci-lint run

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

## snapshot: Build snapshot release (for testing)
snapshot:
	goreleaser release --snapshot --clean

## release: Create a release (requires tag)
release:
	goreleaser release --clean

## fmt: Format code
fmt:
	go fmt ./...

## tidy: Tidy dependencies
tidy:
	go mod tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

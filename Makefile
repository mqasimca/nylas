.PHONY: build test clean install lint

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/mqasimca/nylas/internal/cli.Version=$(VERSION) -X github.com/mqasimca/nylas/internal/cli.Commit=$(COMMIT) -X github.com/mqasimca/nylas/internal/cli.BuildDate=$(BUILD_DATE)"

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/nylas ./cmd/nylas

test:
	go test ./... -v

test-coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

install: build
	cp bin/nylas $(GOPATH)/bin/nylas

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download

# Run tests 3 times as requested
test-3x:
	@echo "=== Test Run 1 of 3 ==="
	go test ./... -v
	@echo ""
	@echo "=== Test Run 2 of 3 ==="
	go test ./... -v
	@echo ""
	@echo "=== Test Run 3 of 3 ==="
	go test ./... -v
	@echo ""
	@echo "=== All 3 test runs completed successfully ==="


# Linting
GOLANGCI_LINT_VERSION=1.50.1

# Build a binary
.PHONY: build
build: CMD = ./cmd/dependr
build:
	go build $(CMD)

# Run test suite
.PHONY: test
test:
	go test -v ./...

# The linting gods must be obeyed
.PHONY: lint
lint: ./bin/golangci-lint
	./bin/golangci-lint run

./bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$(GOLANGCI_LINT_VERSION)

# Generate the mocks (embedded via go generate)
.PHONY: mocks
mocks:
	go generate ./...

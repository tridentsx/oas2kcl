# Project variables
BINARY_NAME = openapi-to-kcl
MAIN_FILE = cmd/main.go

# Default target
.PHONY: all
all: build

# Install dependencies
.PHONY: install
install:
	@echo "Installing dependencies..."
	go mod tidy

# Build the Go binary
.PHONY: build
build: install
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)

# Run the program with an example OpenAPI file
.PHONY: run
run:
	@echo "Running $(BINARY_NAME) with example.json..."
	./$(BINARY_NAME) -oas example.json

# Run the program and generate a .k file
.PHONY: run-out
run-out:
	@echo "Running $(BINARY_NAME) and saving output to schema.k..."
	./$(BINARY_NAME) -oas example.json -out schema.k

# Format the Go code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint the Go code (requires `golangci-lint` to be installed)
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run || echo "Linting issues found"

# Run tests (if you add them)
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Clean the built files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME) schema.k

# Display help
.PHONY: help
help:
	@echo "Makefile commands:"
	@echo "  make install   - Install dependencies"
	@echo "  make build     - Build the project"
	@echo "  make run       - Run the project with example.json"
	@echo "  make run-out   - Run the project and output to schema.k"
	@echo "  make fmt       - Format the code"
	@echo "  make lint      - Lint the code"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Remove built artifacts"


# Project variables
BINARY_NAME = oas2kcl
MAIN_FILE = main.go
OUTPUT_DIR = third_party_licenses


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

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Clean the built files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME) schema.k
	rm -rf $(OUTPUT_DIR)

# License Compliance Check
.PHONY: check-licenses
check-licenses:
	@echo "Checking license compliance for version: $$(cat version.txt)"
	@go-licenses csv . 2>/dev/null | grep -E 'GPL|AGPL|LGPL' && echo "🚨 Non-compliant license detected! Please review dependencies." || echo "✅ All dependencies are compatible."

# Generate License Attributions
.PHONY: generate-attributions
generate-attributions:
	@echo "Generating attribution files..."
	@rm -rf $(OUTPUT_DIR) || true
	@mkdir -p $(OUTPUT_DIR)
	@go-licenses save . --save_path=$(OUTPUT_DIR) --force
	@echo "✅ Attribution files saved in $(OUTPUT_DIR)/"

# Display help
.PHONY: help
help:
	@echo "Makefile commands:"
	@echo "  make install               - Install dependencies"
	@echo "  make build                 - Build the project"
	@echo "  make run                   - Run the project with example.json"
	@echo "  make run-out               - Run the project and output to schema.k"
	@echo "  make fmt                   - Format the code"
	@echo "  make lint                  - Lint the code"
	@echo "  make test                  - Run tests"
	@echo "  make clean                 - Remove built artifacts"
	@echo "  make check-licenses        - Verify all dependencies comply with Apache 2.0"
	@echo "  make generate-attributions - Generate license attribution files in licenses/"


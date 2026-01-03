.PHONY: build generator processor clean clean-logs clean-offsets clean-all run-generator run-processor test bench help

# Build settings
BINARY_DIR := bin
GO := go

# Default target
all: build

# Build all binaries
build:
	@echo "Building..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/generator ./cmd/generator
	$(GO) build -o $(BINARY_DIR)/processor ./cmd/processor
	@echo "Build complete"

# Build individual binaries
generator:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/generator ./cmd/generator

processor:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/processor ./cmd/processor

# Run commands
run-generator: generator
	./$(BINARY_DIR)/generator

run-processor: processor
	./$(BINARY_DIR)/processor

# Generate logs for testing
generate-test-logs: generator
	./$(BINARY_DIR)/generator -count 10000 -interval 1ms

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	$(GO) clean

clean-logs:
	@echo "Cleaning log files..."
	./scripts/clean.sh logs

clean-offsets:
	@echo "Cleaning offset files..."
	./scripts/clean.sh offsets

clean-all:
	@echo "Cleaning everything..."
	./scripts/clean.sh all
	rm -rf $(BINARY_DIR)

# Reset for fresh testing
reset: clean-all build
	@echo "Reset complete - ready for fresh testing"

# Testing
test:
	$(GO) test -v ./...

bench:
	$(GO) test -bench=. -benchmem ./internal/logger/

# Dependencies
deps:
	$(GO) mod tidy
	$(GO) mod download

# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go generate ./...
	@tailwindcss -i internal/web/styles/input.css -o internal/web/assets/css/output.css --minify
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Live Reload
dev:
	@echo "Watching..."
	@air

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main


.PHONY: all build run test clean watch

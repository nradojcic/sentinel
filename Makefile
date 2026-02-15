# Variables
BINARY_NAME=sentinel
BIN_DIR=bin
PROTO_DIR=proto

.PHONY: all build test coverage clean help gen install-tools

all: gen test build

## build: Compiles the binary into the bin directory
build:
	@echo "Building..."
	go build -o $(BIN_DIR)/$(BINARY_NAME) main.go

## test: Runs all tests with the race detector enabled
test:
	@echo "Running tests..."
	go test -v -race ./...

## coverage: Runs tests and generates a coverage report
coverage:
	@echo "Generating coverage report (excluding proto)..."
	go test -coverprofile=coverage.out ./...
	@grep -v "/proto/" coverage.out > coverage.tmp && mv coverage.tmp coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Done! Open coverage.html to see results."

## gen: Generate Go code from Protocol Buffer files
gen:
	@echo "Generating gRPC code..."
	protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           $(PROTO_DIR)/*.proto

## install-tools: Install required protoc plugins
install-tools:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

## clean: Removes the compiled binary and build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
	rm -f $(PROTO_DIR)/*.pb.go
	go clean

## help: Shows this help message
help:
	@echo "Usage: make [target]"
	@grep -E '^##' Makefile | sed -e 's/## //g' | column -t -s ':' | sed -e 's/^/  /'

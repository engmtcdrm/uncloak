.PHONY: menv test testv testcover testcoverall build run lint

PARENT_DIR := $(notdir $(CURDIR))

menv:
	@echo "Current directory: $(CURDIR)"
	@echo "Parent directory name: $(PARENT_DIR)"

test:
	@go test ./...

testv:
	@go test -v ./...

testcover:
	@go test -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html && rm coverage.out

testcoverall:
	@go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html && rm coverage.out

# Build for application
build:
	@echo "Size before build:"; \
	ls -la |grep $(PARENT_DIR); \
	ls -lh |grep $(PARENT_DIR); \
	echo "\n\nSize after build:"; \
	CGO_ENABLED=0 go build --ldflags "-s -w"; \
	strip $(PARENT_DIR); \
	ls -la |grep $(PARENT_DIR); \
	ls -lh |grep $(PARENT_DIR)

# Run for application
run:
	@go run . -t main

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout 5m

.PHONY: menv test testv testcover testcoverall build run buildexample runexample

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
	@go run .

# Build for package examples
buildexample:
	@cd examples; \
	echo "Size before build:"; \
	ls -la |grep examples; \
	ls -lh |grep examples; \
	echo "\n\nSize after build:"; \
	CGO_ENABLED=0 go build --ldflags "-s -w"; \
	strip examples; \
	ls -la |grep examples; \
	ls -lh |grep examples; \
	cd ..

# Run for package examples
runexample:
	@cd examples; \
	go run .; \
	cd ..

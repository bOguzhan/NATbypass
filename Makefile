.PHONY: all build test clean run-mediatory run-application test-config

GO=go
BUILD_DIR=bin

all: build

build: build-mediatory build-application build-stun-test

build-mediatory:
	@echo "Building mediatory server..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/mediatory-server cmd/mediatory-server/main.go

build-application:
	@echo "Building application server..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/application-server cmd/application-server/main.go

build-stun-test:
	@echo "Building STUN test..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/stun_test test/main.go

test:
	@echo "Running all tests..."
	$(GO) test -v ./...

test-config:
	@echo "Testing configuration system..."
	$(GO) test -v ./test/config

test-networking: 
	@echo "Testing networking package..."
	$(GO) test -v ./pkg/networking

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)

run-mediatory: build-mediatory
	@echo "Running mediatory server..."
	./$(BUILD_DIR)/mediatory-server

run-application: build-application
	@echo "Running application server..."
	./$(BUILD_DIR)/application-server


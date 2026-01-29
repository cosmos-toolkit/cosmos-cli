.PHONY: build run test clean install

BINARY_NAME := cosmos
BUILD_DIR := bin
MAIN_PATH := ./cmd/cosmos

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	go install $(MAIN_PATH)

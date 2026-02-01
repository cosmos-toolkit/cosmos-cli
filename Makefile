.PHONY: build run test clean install install-remote release

BINARY_NAME := cosmos
BUILD_DIR := bin
RELEASE_DIR := release
MAIN_PATH := ./cmd/cosmos
# Full module path for "install like Docker" (go install from anywhere)
INSTALL_PATH := github.com/cosmos-toolkit/cosmos-cli/cmd/cosmos

# Version for release tarballs (set via: make release VERSION=1.0.0)
VERSION ?= dev

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR) $(RELEASE_DIR)

# Install binary to $HOME/go/bin (or GOPATH/bin). Requires that dir in PATH.
install:
	go install $(MAIN_PATH)

# Same as install but fetches from remote (no clone needed). For README/CI.
install-remote:
	go install $(INSTALL_PATH)@latest

# Build binaries for all platforms (for GitHub Releases). Creates release/cosmos-<VERSION>-<os>-<arch>.tar.gz
# Usage: make release VERSION=1.0.0
release:
	@mkdir -p $(RELEASE_DIR)
	@for goos in darwin linux; do \
		for goarch in amd64 arm64; do \
			GOOS=$$goos GOARCH=$$goarch go build -o $(RELEASE_DIR)/$(BINARY_NAME) $(MAIN_PATH) && \
			(cd $(RELEASE_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-$$goos-$$goarch.tar.gz $(BINARY_NAME) && rm -f $(BINARY_NAME)); \
		done; \
	done
	@GOOS=windows GOARCH=amd64 go build -o $(RELEASE_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@(cd $(RELEASE_DIR) && zip -q $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME).exe && rm -f $(BINARY_NAME).exe)
	@echo "Release artifacts in $(RELEASE_DIR)/"
	@ls -la $(RELEASE_DIR)/

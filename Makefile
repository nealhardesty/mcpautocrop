BINARY_NAME   := mcpautocrop
VERSION       := $(shell grep 'Version = ' version.go | cut -d'"' -f2)
BUILD_DIR     := dist
TESTDATA_DIR  := testdata
TESTIMG_URL   := https://raw.githubusercontent.com/golang/image/master/testdata/bw-gopher.png
TESTIMG_SRC   := $(TESTDATA_DIR)/sample.png
TESTIMG_OUT   := $(TESTDATA_DIR)/sample_cropped.png

.PHONY: help build test run clean lint fmt tidy version version-increment release push \
        fetch-testimg demo

## help: Show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | \
		awk -F': ' '{printf "  %-22s %s\n", substr($$1, 4), $$2}'

## build: Compile the binary
build:
	@go build -ldflags="-s -w" -o $(BINARY_NAME) .

## test: Run all unit tests with race detector
test:
	@go test -race -v ./...

## fetch-testimg: Download a sample PNG into testdata/
fetch-testimg:
	@mkdir -p $(TESTDATA_DIR)
	@echo "Downloading sample image..."
	@curl -fsSL -o $(TESTIMG_SRC) "$(TESTIMG_URL)"
	@echo "Saved to $(TESTIMG_SRC)"

## demo: Fetch sample image, build, and run a live crop (testdata/ required)
demo: fetch-testimg build
	@echo "Running: ./$(BINARY_NAME) test $(TESTIMG_SRC) $(TESTIMG_OUT)"
	@./$(BINARY_NAME) test $(TESTIMG_SRC) $(TESTIMG_OUT)

## run: Build and run the MCP server (stdio)
run: build
	@./$(BINARY_NAME) mcp

## clean: Remove build artifacts and generated test outputs
clean:
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f $(TESTIMG_OUT)

## lint: Run linters
lint:
	@go vet ./...
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

## fmt: Format code
fmt:
	@gofmt -w .
	@which goimports >/dev/null 2>&1 && goimports -w . || true

## tidy: Run go mod tidy
tidy:
	@go mod tidy

## version: Display current version
version:
	@echo $(VERSION)

## version-increment: Bump the patch version in version.go
version-increment:
	@current=$(VERSION); \
	major=$$(echo $$current | cut -d. -f1); \
	minor=$$(echo $$current | cut -d. -f2); \
	patch=$$(echo $$current | cut -d. -f3); \
	next=$$((patch + 1)); \
	new_version="$$major.$$minor.$$next"; \
	sed -i "s/Version = \"$$current\"/Version = \"$$new_version\"/" version.go; \
	echo "Version bumped: $$current -> $$new_version"

## release: Build release binaries for Linux, macOS, and Windows
release: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64   .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64  .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Release binaries written to $(BUILD_DIR)/"

## push: Version bump, build, commit, tag, and push
push: tidy fmt test build
	@$(MAKE) version-increment
	@new_version=$$(grep 'Version = ' version.go | cut -d'"' -f2); \
	git add -A; \
	git commit -m "release: v$$new_version"; \
	git tag "v$$new_version"; \
	git push && git push --tags
	@echo "Released v$$(grep 'Version = ' version.go | cut -d'"' -f2)"

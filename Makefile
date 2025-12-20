all: build

# Version and application information
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
REPO := ghcr.io/silenceper/mcp-k8s
APPNAME := mcp-k8s
BUILDDIR := ./bin

# Build flags
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/mcp-k8s cmd/mcp-k8s/main.go

# Clean the project
.PHONY: clean
clean:
	rm -rf $(BUILDDIR)

# Format the code
.PHONY: fmt
fmt:
	go fmt ./...

# Create output directory
.PHONY: init
init:
	mkdir -p $(BUILDDIR)
# Cross-platform build - Windows
.PHONY: build-windows-amd64
build-windows-amd64: init
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/$(APPNAME)_windows_amd64.exe  cmd/mcp-k8s/main.go

# Cross-platform build - macOS (Intel)
.PHONY: build-darwin-amd64
build-darwin-amd64: init
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/$(APPNAME)_darwin_amd64  cmd/mcp-k8s/main.go

# Cross-platform build - macOS (Apple Silicon)
.PHONY: build-darwin-arm64
build-darwin-arm64: init
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/$(APPNAME)_darwin_arm64  cmd/mcp-k8s/main.go

# Cross-platform build - Linux
.PHONY: build-linux-amd64
build-linux-amd64: init
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/$(APPNAME)_linux_amd64  cmd/mcp-k8s/main.go

# Cross-platform build - Linux
.PHONY: build-linux-arm64
build-linux-arm64: init
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/$(APPNAME)_linux_arm64  cmd/mcp-k8s/main.go

# Cross-platform build - All platforms
.PHONY: build-all
build-all: build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64
	@echo "All platforms built successfully"
	@ls -la $(BUILDDIR)

.PHONY: docker-build
docker-build:
	docker build -t $(REPO):$(VERSION) -f Dockerfile .

.PHONY: docker-push
docker-push:
	docker push $(REPO):$(VERSION)
all: build
.PHONY: build
build:
	go build -o ./bin/mcp-k8s cmd/server/main.go


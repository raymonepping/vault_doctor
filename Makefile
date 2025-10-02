APP := $(shell basename $(CURDIR))
VERSION ?= dev
GOFLAGS := -trimpath
LDFLAGS := -s -w -X 'main.Version=$(VERSION)'

.PHONY: all build fmt vet lint run test clean release

all: build

fmt:
	@go fmt ./...

vet:
	@go vet ./...

# Uncomment if you have golangci-lint installed
# lint:
#	@golangci-lint run ./...

build: fmt vet
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(APP) ./cmd/$(APP)

run: build
	@./bin/$(APP)

test:
	@go test ./...

clean:
	@rm -rf bin dist

# Cross-platform release builds
release: clean fmt vet
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP)_$(VERSION)_darwin_amd64 ./cmd/$(APP)
	GOOS=darwin GOARCH=arm64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP)_$(VERSION)_darwin_arm64 ./cmd/$(APP)
	GOOS=linux  GOARCH=amd64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP)_$(VERSION)_linux_amd64 ./cmd/$(APP)
	GOOS=linux  GOARCH=arm64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(APP)_$(VERSION)_linux_arm64 ./cmd/$(APP)
	@echo "Built release artifacts in ./dist:"
	@ls -1 dist

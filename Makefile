BUILD_OS            := windows linux darwin
BUILD_ARCH          := amd64
BUILD_DEBUG_FLAGS   :=
BUILD_RELEASE_FLAGS := -ldflags "-s -w"

BUILD_PATH          := bin

CURRENT_OS          := $(shell uname | tr A-Z a-z)
CURRENT_ARCH        := amd64

SOURCES             := $(shell find . -name '*.go' -not -path './vendor/*')
PACKAGES            := $(sort $(dir $(SOURCES)))
BINARIES            := $(notdir $(shell find ./cmd -mindepth 1 -maxdepth 1 -type d))

DOCKER_REPO			:= arsperger

build: bin
	CGO_ENABLED=0 GOOS=$(CURRENT_OS) GOARCH=$(BUILD_ARCH) go build -o $(BUILD_PATH)/$(BINARIES) $(BUILD_RELEASE_FLAGS) $(SOURCES)

docker: Dockerfile
	docker build -t arsperger/ws-registrar .

.PHONY: build docker
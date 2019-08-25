PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
LDFLAGS := $(shell echo "")

GOOS := $(shell go env GOOS)
GOOSALT ?= 'linux'
ifeq ($(GOOS),'darwin')
  GOOSALT = 'mac'
endif

BUILD_LDFLAGS := '$(LDFLAGS)'

all: build

build:
	@echo "building pong binary to ./pong"
	@GOPROXY=https://proxy.golang.org GO111MODULE=on GO_FLAGS="" CGO_ENABLED=0 go build -tags kqueue --ldflags $(BUILD_LDFLAGS)

build-linux:
	@echo "building pong-linux-amd64 to ./pong-linux-amd64"
	@GOOS=linux GOARCH=amd64 GOPROXY=https://proxy.golang.org GO111MODULE=on GO_FLAGS="" CGO_ENABLED=0 go build -o pong-linux-amd64 -tags kqueue --ldflags $(BUILD_LDFLAGS)

build-darwin:
	@echo "building pong-linux-amd64 to ./pong-darwin-amd64"
	@GOOS=darwin GOARCH=amd64 GOPROXY=https://proxy.golang.org GO111MODULE=on GO_FLAGS="" CGO_ENABLED=0 go build -o pong-darwin-amd64 -tags kqueue --ldflags $(BUILD_LDFLAGS)

build-windows:
	@echo "building pong-linux-amd64 to ./pong-windows-amd64"
	@GOOS=windows GOARCH=amd64 GOPROXY=https://proxy.golang.org GO111MODULE=on GO_FLAGS="" CGO_ENABLED=0 go build -o pong-windows-amd64 -tags kqueue --ldflags $(BUILD_LDFLAGS)

release: build-linux build-darwin build-windows
	@echo "releasing multi-platform pong binaries to releases/"
	@mkdir -p releases/
	@mv pong-linux-amd64	releases/pong-linux-amd64
	@mv pong-windows-amd64	releases/pong-windows-amd64
	@mv pong-darwin-amd64	releases/pong-darwin-amd64
	@tar cvzf pong-multiplatform.tar.gz releases/

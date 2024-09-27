# Default test path and formatting
TEST ?= ./...
GOFMT_FILES ?= $(shell find . -name '*.go' |grep -v vendor)
OPENTOFU_PATH ?= $(shell which tofu)

# Extract binary name from go.mod
BINARY_NAME?=$(shell grep "^module" go.mod | awk -F'/' '{print $$NF}')

# Optional version tag for binary
BINARY_VERSION ?= ''

# Default build parameters
PARAM_CC ?= musl-gcc
PARAM_GOOS ?= linux
PARAM_GOARCH ?= amd64
PARAM_CGO_ENABLED ?= 1
PARAM_CGO_LDFLAGS ?= '-s -w -static -Wl,-unresolved-symbols=ignore-all'
PARAM_VERIFY ?= 'statically linked'

# Setup default environment variables
.PHONY: set-env
set-env:
	go env -w GOOS=$(PARAM_GOOS)
	go env -w GOARCH=$(PARAM_GOARCH)
	go env -w CGO_ENABLED=$(PARAM_CGO_ENABLED)

# Setup linux environment variables
.PHONY: set-env-linux
set-env-linux: set-env
	go env -w CC=$(PARAM_CC)
	go env -w CGO_LDFLAGS=$(PARAM_CGO_LDFLAGS)

# Setup windows environment variables
.PHONY: set-env-windows
set-env-windows: set-env
	go env -w CC=/usr/bin/x86_64-w64-mingw32-gcc
	go env -w CGO_LDFLAGS=$(PARAM_CGO_LDFLAGS)

# Generate documentation on linux
.PHONY: create-docs-linux
create-docs-linux: set-env-linux
	terraform fmt -recursive ./examples/
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

# Build a static linux binary
.PHONY: build-linux
build-linux: set-env-linux
	go build -v -o $(BINARY_NAME)$(BINARY_VERSION) .

# Build statically binary for linux amd64
.PHONY: build-linux-amd64
build-linux-amd64:
	$(MAKE) build-linux

# Build statically binary for linux arm64
.PHONY: build-linux-arm64
build-linux-arm64:
	$(MAKE) PARAM_GOARCH="arm64" build-linux

# Build a macOS binary
.PHONY: build-darwin
build-darwin: set-env
	go build -v -o $(BINARY_NAME)$(BINARY_VERSION) .

# Build a macOS binary for amd64
.PHONY: build-darwin-amd64
build-darwin-amd64:
	$(MAKE) PARAM_GOOS="darwin" build-darwin

# Build a macOS binary for arm64
.PHONY: build-darwin-arm64
build-darwin-arm64:
	$(MAKE) PARAM_GOOS="darwin" PARAM_GOARCH="arm64" build-darwin

# Build a Windows binary
.PHONY: build-windows
build-windows: set-env-windows
	go build -v -o $(BINARY_NAME)$(BINARY_VERSION).exe .

# Build a Windows binary for amd64
.PHONY: build-windows-amd64
build-windows-amd64:
	$(MAKE) PARAM_GOOS="windows" build-windows

# Verify binary
.PHONY: verify-binary
verify-binary:
	file $(BINARY_NAME)$(BINARY_VERSION) | grep -i $(PARAM_VERIFY)

# Verify linux binary
.PHONY: verify-binary-linux
verify-binary-linux:
	$(MAKE) verify-binary

# Verify macOS binary amd64
.PHONY: verify-binary-darwin-amd64
verify-binary-darwin-amd64:
	$(MAKE) PARAM_VERIFY="'executable x86_64'" verify-binary

# Verify macOS binary arm64
.PHONY: verify-binary-darwin-arm64
verify-binary-darwin-arm64:
	$(MAKE) PARAM_VERIFY="'executable arm64'" verify-binary

# Verify Windows binary amd64
.PHONY: verify-binary-windows-amd64
verify-binary-windows-amd64:
	file $(BINARY_NAME)$(BINARY_VERSION).exe | grep -i "PE32+ executable (console) x86-64"

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 \
	go test $(TEST) -v $(TESTARGS) -timeout 10m

# Run unit tests
.PHONY: test
test:
	go test $(TEST) -timeout=30s -parallel=4

# Run gofmt on all Go files
.PHONY: fmt
fmt:
	gofmt -w $(GOFMT_FILES)

# Run acceptance tests with OpenTofu
.PHONY: testacc_tofu
testacc_tofu:
	TF_ACC=1 \
	TF_ACC_TERRAFORM_PATH=$(OPENTOFU_PATH) \
	TF_ACC_PROVIDER_NAMESPACE="hashicorp" \
	TF_ACC_PROVIDER_HOST="registry.opentofu.org" \
	go test $(TEST) -v $(TESTARGS) -timeout 10m

# Reset local go env
.PHONY: reset-go-env
reset-go-env:
	go env -u GOROOT
	go env -u GOPATH
	go env -u GOARCH
	go env -u GOOS
	go env -u CGO_ENABLED
	go env -u GO111MODULE
	go env -u CGO_LDFLAGS

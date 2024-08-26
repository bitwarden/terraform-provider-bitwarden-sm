TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
OPENTOFU_PATH?=$$(which tofu)

# Extract binary name from go.mod
BINARY_NAME?=$$(grep "^module" go.mod | awk -F'/' '{print $$NF}')

# Build parameters
PARAM_CC?=musl-gcc
PARAM_GOOS?=linux
PARAM_GOARCH?=amd64
PARAM_CGO_ENABLED?=1
PARAM_CGO_LDFLAGS?='-static -Wl,-unresolved-symbols=ignore-all'
PARAM_VERIFY?='static'

# Build a static linux binary
.PHONY: build-linux
build-linux:
	CC=$(PARAM_CC) \
  	GOOS=$(PARAM_GOOS) \
  	GOARCH=$(PARAM_GOARCH) \
	CGO_ENABLED=$(PARAM_CGO_ENABLED) \
  	CGO_LDFLAGS=$(PARAM_CGO_LDFLAGS) \
  	go build -v -o $(BINARY_NAME) .

# Build statically binary for linux amd64
.PHONY: build-linux-amd64
build-linux-amd64:
	$(MAKE) build-linux

# Build statically binary for linux arm64
.PHONY: build-linux-arm64
build-linux-arm64:
	$(MAKE) PARAM_GOARCH="arm64" build-linux

# Build a macOS binary
.PHONY: build-macos
build-macos:
	GOOS=$(PARAM_GOOS) \
	GOARCH=$(PARAM_GOARCH) \
	CGO_ENABLED=$(PARAM_CGO_ENABLED) \
	go build -v -o $(BINARY_NAME) .

# Build a macOS binary for amd64
.PHONY: build-macos-amd64
build-macos-amd64:
	$(MAKE) PARAM_GOOS="darwin" build-macos

# Build a macOS binary for arm64
.PHONY: build-macos-arm64
build-macos-arm64:
	$(MAKE) PARAM_GOOS="darwin" PARAM_GOARCH="arm64" build-macos

# Verify binary
.PHONY: verify-binary
verify-binary:
	file $(BINARY_NAME) | grep -i $(PARAM_VERIFY)

# Verify linux binary
.PHONY: verify-binary-linux
verify-binary-linux:
	$(MAKE) verify-binary

# Verify macos binary amd64
.PHONY: verify-binary-macos-amd64
verify-binary-macos-amd64:
	$(MAKE) PARAM_VERIFY='executable x86_64' verify-binary

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


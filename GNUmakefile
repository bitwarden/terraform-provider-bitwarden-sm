TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
OPENTOFU_PATH?=$$(which tofu)

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 10m

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

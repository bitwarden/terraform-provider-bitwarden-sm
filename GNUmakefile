TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

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

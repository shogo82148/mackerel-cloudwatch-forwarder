help: ## Show this text.
	# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

VERSION=$(patsubst "%",%,$(lastword $(shell grep 'const Version' version.go)))
ARTIFACTS_DIR=$(CURDIR)/artifacts/$(VERSION)
RELEASE_DIR=$(CURDIR)/release/$(VERSION)
LATEST_DIR=$(CURDIR)/release/latest
SRC_FILES=$(shell find . -type f -name '*.go')
GO111MODULE=on

.PHONY: all clean build test

all: $(LATEST_DIR)/mackerel-cloudwatch-forwarder.zip ## Build all binaries
build: $(ARTIFACTS_DIR)/mackerel-cloudwatch-forwarder ## Build executable binary

clean: ## Remove built files.
	rm -rf $(CURDIR)/artifacts
	rm -rf $(CURDIR)/release

test: ## Run test.
	go test -v -race ./...
	go vet ./...

$(ARTIFACTS_DIR):
	mkdir -p $@

$(ARTIFACTS_DIR)/mackerel-cloudwatch-forwarder: $(ARTIFACTS_DIR) $(SRC_FILES) go.mod go.sum
	./run-in-docker.sh go build -o $@ ./cmd/mackerel-cloudwatch-forwarder

$(RELEASE_DIR):
	mkdir -p $@

$(LATEST_DIR):
	mkdir -p $@

$(RELEASE_DIR)/mackerel-cloudwatch-forwarder.zip: $(RELEASE_DIR) $(ARTIFACTS_DIR)/mackerel-cloudwatch-forwarder
	cd $(ARTIFACTS_DIR) && zip -9 $@ *

$(LATEST_DIR)/mackerel-cloudwatch-forwarder.zip: $(LATEST_DIR) $(RELEASE_DIR)/mackerel-cloudwatch-forwarder.zip
	cp $(RELEASE_DIR)/mackerel-cloudwatch-forwarder.zip $@

.PHONY: release-sam

release-sam: $(LATEST_DIR)/mackerel-cloudwatch-forwarder.zip template.yaml ## Release the application to AWS Serverless Application Repository
	sam package \
		--template-file template.yaml \
		--output-template-file packaged.yaml \
		--s3-bucket shogo82148-sam
	sam publish \
		--template packaged.yaml \
		--region us-east-1

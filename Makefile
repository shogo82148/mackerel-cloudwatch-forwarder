help: ## Show this text.
	# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

SRC_FILES=$(shell find . -type f -name '*.go')

.PHONY: all clean build test

all: dist.zip template.yaml ## Build all binaries
build: dist/bootstrap ## Build executable binary

clean: ## Remove built files.
	-rm -f dist.zip
	-rm -rf dist
	-docker volume rm mackerel-cloudwatch-forwarder-cache

test: ## Run test.
	go test -v -race ./...
	go vet ./...

dist/bootstrap: $(SRC_FILES) go.mod go.sum
	mkdir -p dist
	./run-in-docker.sh go build -tags lambda.norpc -o dist/bootstrap ./cmd/mackerel-cloudwatch-forwarder

dist.zip: dist/bootstrap
	cd dist && zip -r ../dist.zip .

version.go template.yaml: VERSION generate.sh template.template.yaml
	./generate.sh

.PHONY: release-sam
release-sam: dist.zip template.yaml ## Release the application to AWS Serverless Application Repository
	sam package \
		--template-file template.yaml \
		--output-template-file packaged.yaml \
		--s3-bucket shogo82148-sam
	sam publish \
		--template packaged.yaml \
		--region us-east-1

SHELL:=/bin/bash
-include .env

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: ## Build the api
	go build .

.PHONY: test
test: ## Test the api
	go test -v ./...

.PHONY: run
run: ## Run the api
	UBERBASE_IMAGE_PATH=./images go run .

.PHONY: image
image: ## Build the base containerd image
	docker build -t bluetongueai/functions-hello-world:latest -f ../images/hello-world/Dockerfile .

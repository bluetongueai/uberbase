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
build:
	mkdir -p bin
	GOOS=linux GOARCH=$(go env GOARCH) go build -buildvcs=false -o bin/uberbase .
	chmod +x bin/uberbase

.PHONY: install
install:
	cp bin/uberbase /usr/bin/.

.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	go run .

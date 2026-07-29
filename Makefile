SHELL := /bin/bash

GO ?= go
NPM ?= npm

export GOCACHE := $(CURDIR)/.cache/go-build
export GOMODCACHE := $(CURDIR)/.cache/go-mod
export GOPATH := $(CURDIR)/.cache/go-path

.PHONY: bootstrap build test test-offline lint fmt clean sandbox sandbox-down sandbox-shell

bootstrap:
	@mkdir -p .data .cache .artifacts .coverage bin

build: bootstrap
	$(GO) build ./...

test: bootstrap
	$(GO) test ./...

test-offline:
	docker compose --profile offline run --rm offline-test

lint:
	$(GO) vet ./...

fmt:
	@test -z "$$($(GO) fmt ./...)"

clean:
	$(GO) clean ./...

sandbox:
	docker compose up -d dev

sandbox-down:
	docker compose down

sandbox-shell:
	docker compose exec dev bash

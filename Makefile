#!/usr/bin/make -f

VERSION := $(shell git describe)

test: fmt
	go test -short -count=200 -race -cover -timeout=1s ./...

fmt:
	@go version && go fmt ./... && go mod tidy

test.load: test
	go test -v -race -cover -run=TestLoad

.PHONY: test fmt test.load

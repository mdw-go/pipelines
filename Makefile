#!/usr/bin/make -f

VERSION := $(shell git describe)

test: fmt
	go test -count 200 -race -cover -timeout=1s -count=1 ./...

fmt:
	@go version && go fmt ./... && go mod tidy

.PHONY: deps proto-gen help grpc-ui

APP_VERSION := $(shell git rev-parse --short HEAD || echo "GitNotFound")

vars:
	@echo APP_VERSION=$(APP_VERSION)

help:
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' ./Makefile | awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

deps:
	go mod tidy
	go mod download

proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false pkg/command/command.proto

run: vars
	go run cmd/server.go

build-local: vars
	go build -o cmd/server cmd/server.go

grpc-ui:
	grpcui -plaintext localhost:3009


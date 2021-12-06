.PHONY: deps proto

deps:
	go mod tidy
	go mod vendor

proto_gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative command/command.proto

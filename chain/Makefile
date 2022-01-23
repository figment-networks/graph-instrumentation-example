.PHONY: build proto generate

# Build the binary
build:
	go build

# Generate protobuf package code
proto:
	protoc \
		--go_out=paths=source_relative:. \
		./proto/codec.proto

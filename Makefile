bin_dir=$(shell pwd)/bin

bin/protoc-gen-go: go.mod
	GOBIN=$(bin_dir) go install google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: generate
generate: bin/protoc-gen-go
	PATH="$(bin_dir):$(PATH)" protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		rpc/document.proto \
		&& go run cmd/codegen/main.go

.PHONY: test
test:
	go test -short -v ./...
	golangci-lint run ./...

.PHONY: test-race
test-race:
	go test -short -race ./...


GOPATH := $(CURDIR)/../.gopath
GOCACHE := $(CURDIR)/../.gocache

.PHONY: test server cli tidy proto install-tools setup check-protoc

test:
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go test ./...

server:
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go run ./cmd/server

cli:
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go run ./cmd/cli

tidy:
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go mod tidy

proto: check-protoc
	PATH=$(GOPATH)/bin:$$PATH protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/vitals/v1/vitals.proto

check-protoc:
	@which protoc > /dev/null || (echo "Error: protoc not found. Install it with:\n  macOS: brew install protobuf\n  Linux: sudo apt install protobuf-compiler" && exit 1)

install-tools:
	@echo "Installing Go protoc plugins..."
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOPATH=$(GOPATH) GOCACHE=$(GOCACHE) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Done! Make sure $(GOPATH)/bin is in your PATH"

setup: install-tools
	@echo "Setup complete! Run 'make proto' to generate protobuf code."

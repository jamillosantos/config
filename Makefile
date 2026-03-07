
.PHONY: mod
mod:
	go mod vendor -v

.PHONY: lint
lint:
	go tool golangci-lint run --timeout 5m

.PHONY: test
test: lint
	go test ./...

.PHONY: all
all: lint sec test
	@:

.PHONY: generate
generate:
	go generate ./...

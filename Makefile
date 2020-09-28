.PHONY: dev build check lint

dev: build

build:
	@echo "Building...."
	@go build -o ./dist/bin/api ./cmd/api

setup-lint:
	@echo "Installing dev dependencies...."
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.20.0

lint:
	@echo "Linting...."
	@golangci-lint run --issues-exit-code 1 -v ./...

check:
	@echo "Testing...."
	@go test -v \
		-bench . -benchmem \
		-cover -covermode count \
		./...

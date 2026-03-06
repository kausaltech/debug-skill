.PHONY: build test lint all clean

all: lint test build

build:
	go build -o dap ./cmd/dap

test:
	go test ./...

lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

clean:
	rm -f dap

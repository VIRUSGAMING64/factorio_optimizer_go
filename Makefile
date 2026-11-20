.PHONY: test test-v build run-rg run-six

test:
	go test ./...

test-v:
	go test ./... -count=1 -v

build:
	go build -o bin/optimize ./cmd/optimize

run-rg: build
	./bin/optimize --config=data/red_green.json --sa-iters=30000

run-six: build
	./bin/optimize --config=data/config.json --sa-iters=8000

.PHONY: run test build clean science blue

run:
	go run . -config=data/red_green.json -sa-iters=20000

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/optimize .

clean:
	rm -rf bin solves

science:
	go run . -config=data/config.json -sa-iters=15000

blue:
	go run . -config=data/blue_circuits.json -sa-iters=25000

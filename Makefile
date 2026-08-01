.PHONY: build run test bench clean

build:
	go build -o bin/distlimiter-server ./cmd/server
	go build -o bin/distlimiter-bench ./cmd/benchmark

run: build
	./bin/distlimiter-server

test:
	go test -v -race ./internal/...

bench: build
	./bin/distlimiter-bench -n 10000 -c 50

clean:
	rm -rf bin/

.PHONY: build run test

build:
	go build -o bin/bookcabin ./cmd/http

run: build
	./bin/bookcabin

test:
	go test -race ./...
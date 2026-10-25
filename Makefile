.PHONY: build run sprint

build:
	@go build -o bin/api ./cmd/api

run:
	@./bin/api

sprint: build run
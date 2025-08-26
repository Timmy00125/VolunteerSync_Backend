.PHONY: run test gen fmt

gen:
	go run github.com/99designs/gqlgen generate

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

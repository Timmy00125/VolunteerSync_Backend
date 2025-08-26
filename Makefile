.PHONY: run test gen fmt

gen:
	gqlgen generate --verbose

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	go fmt ./...

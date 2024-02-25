run:
	go run cmd/main.go

tools:
	go install github.com/matryer/moq@latest

test:
	go test -v ./...
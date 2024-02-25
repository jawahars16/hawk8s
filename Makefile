run:
	go run cmd/hawk8s/main.go

tools:
	go install github.com/matryer/moq@latest

test:
	go test -v ./...
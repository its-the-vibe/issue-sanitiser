# Makefile for issue-sanitiser Go project

APP_NAME=issue-sanitiser


.PHONY: build run test lint clean install

build:
	go build -o $(APP_NAME) main.go

install:
	go install

run: build
	./$(APP_NAME)

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "Run 'go fmt ./...' to fix formatting"; exit 1)

clean:
	rm -f $(APP_NAME)

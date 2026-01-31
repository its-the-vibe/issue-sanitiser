# Makefile for issue-sanitiser Go project

APP_NAME=issue-sanitiser


.PHONY: build run test clean install

build:
	go build -o $(APP_NAME) main.go

install:
	go install

run: build
	./$(APP_NAME)

test:
	go test ./...

clean:
	rm -f $(APP_NAME)

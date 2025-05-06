.PHONY: build run tidy

build:
	go build -o switch-checker *.go

run: tidy build
	./switch-checker

tidy:
	go mod tidy

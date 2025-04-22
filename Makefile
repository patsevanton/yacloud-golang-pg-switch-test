.PHONY: build run tidy

build:
	go build -o switch-checker main.go config.go dbclient.go checker.go

run: tidy build
	./switch-checker

tidy:
	go mod tidy

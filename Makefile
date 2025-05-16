.PHONY: build run tidy

run:
	go mod tidy
	go build -o switch-checker *.go
	./switch-checker

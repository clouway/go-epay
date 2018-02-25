.DEFAULT_GOAL := build

test:
	go test ./...
build:
	go install ./...


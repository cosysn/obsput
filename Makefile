.PHONY: test build clean

test:
	go test ./...

build:
	go build -o obsput main.go

clean:
	rm -f obsput

.PHONY: all build test clean

all: build

build:
	go build -o trees-cli ./cmd/trees-cli
	go build -o trees-server .

test:
	go test -v ./...

clean:
	rm -f trees-cli trees-server

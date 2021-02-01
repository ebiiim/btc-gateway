.PHONY: all test testlocal

all: test

test:
	go test -race -cover ./...

testlocal:
	go test -race -cover -tags localTest ./...

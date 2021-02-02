.PHONY: all test testlocal

all: test

test:
	go test -race -cover ./...

testlocal:
	go test -count=1 -race -cover -tags localTest ./...

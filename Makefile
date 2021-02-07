include .env
export

.PHONY: all gen build test test-local store-create-dynamodb store-delete-dynamodb

all: test build

gen:
	go generate ./...

build: gen
	go build "-ldflags=-s -w" -trimpath btcgw.go

test: gen
	go test -race -cover ./...

test-local: gen
	go test -count=1 -race -cover -tags localTest ./...

store-create-dynamodb:
	aws dynamodb create-table \
    --table-name anchors.btcgw \
    --attribute-definitions \
        AttributeName=cid,AttributeType=S \
    --key-schema AttributeName=cid,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

store-delete-dynamodb:
	aws dynamodb delete-table --table-name anchors.btcgw

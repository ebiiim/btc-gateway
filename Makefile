include .env
export

.PHONY: all test testlocal

all: test

test:
	go test -race -cover ./...

test-local:
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

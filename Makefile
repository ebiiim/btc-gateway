include .env
export

.PHONY: all gen build test test-local store-create-dynamodb store-delete-dynamodb api-generate-swagger-ui

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

api-generate-swagger-ui:
	mkdir -p vendor
	[ ! -d "swagger-ui" ] && git clone https://github.com/swagger-api/swagger-ui vendor/swagger-ui || echo "vendor/swagger-ui already exists"
	rm -rf dist/swagger-ui && mkdir -p dist/swagger-ui && cp -r vendor/swagger-ui/dist/* dist/swagger-ui
	sed -i -e "s|https://petstore.swagger.io/v2/swagger.json|openapi.yml|g" dist/swagger-ui/index.html
	cp api/openapi.yml dist/swagger-ui/

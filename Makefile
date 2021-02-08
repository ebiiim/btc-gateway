include .env
export

.PHONY: all gen build test test-local store-create-dynamodb store-delete-dynamodb api-generate-swagger-ui deploy-download-bitcoincore

all: test build api-generate-swagger-ui

gen:
	go generate ./...

build: gen
	go build "-ldflags=-s -w" -trimpath btcgw.go

test: gen
	go test -race -cover ./...

## localTest needs a bitcoind server.
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

# please serve dist/swagger-ui
api-generate-swagger-ui:
	mkdir -p tmp
	[ ! -d "tmp/swagger-ui" ] && git clone https://github.com/swagger-api/swagger-ui tmp/swagger-ui || echo "tmp/swagger-ui already exists"
	rm -rf dist/swagger-ui && mkdir -p dist/swagger-ui && cp -r tmp/swagger-ui/dist/* dist/swagger-ui
	sed -i -e "s|https://petstore.swagger.io/v2/swagger.json|openapi.yml|g" dist/swagger-ui/index.html
	cp api/openapi.yml dist/swagger-ui/

# docker build
URL_BITCOIN_CORE="https://bitcoincore.org/bin/bitcoin-core-0.21.0/bitcoin-0.21.0-x86_64-linux-gnu.tar.gz"
ZIP_BITCOIN_CORE="bitcoin-0.21.0-x86_64-linux-gnu.tar.gz"
DIR_BITCOIN_CORE="bitcoin-0.21.0/bin"
deploy-download-bitcoincore:
	mkdir -p tmp
	[ ! -e "tmp/$(ZIP_BITCOIN_CORE)" ] \
	&& curl $(URL_BITCOIN_CORE) -o./tmp/$(ZIP_BITCOIN_CORE) \
	|| echo "tmp/$(ZIP_BITCOIN_CORE) already exists"
	tar -zxvf "tmp/$(ZIP_BITCOIN_CORE)" -C tmp/
	[ -e bitcoin-cli ] && rm bitcoin-cli || echo "bitcoin-cli not found"
	cp "tmp/$(DIR_BITCOIN_CORE)/bitcoin-cli" .
	# stop if wrong arch
	./bitcoin-cli --version

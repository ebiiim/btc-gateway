FROM golang:1.15-buster as builder
WORKDIR /go/src/app
COPY . .
RUN make deploy-download-bitcoincore
RUN go generate ./...
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o main cmd/btcgw/btcgw.go

FROM alpine:3.13
COPY --from=builder /go/src/app/main .
COPY --from=builder /go/src/app/bitcoin-cli .
ENTRYPOINT [ "./main" ]

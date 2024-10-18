FROM golang:1.21.6-alpine3.18 AS Build

WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o /kvs main.go

FROM alpine:latest

WORKDIR /

COPY --from=Build /kvs /kvs
COPY cert.pem /cert.pem
COPY key.pem /key.pem

ENTRYPOINT [ "/kvs" ]

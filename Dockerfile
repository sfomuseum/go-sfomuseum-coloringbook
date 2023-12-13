FROM rust:alpine AS rusttools

RUN cargo install vtracer

FROM golang:1.21-alpine as gotools

RUN mkdir /build
COPY . /build/go-sfomuseum-colouringbook

RUN apk update && apk upgrade \
    && cd /build/go-sfomuseum-colouringbook \
    && go build -mod vendor -ldflags="-s -w" -o /usr/local/bin/pdf cmd/pdf/main.go \
    && cd \
    && rm -rf build
    
FROM alpine

RUN apk update && apk upgrade \
    && apk add openjdk21-jre

COPY --from=rusttools /usr/local/cargo/bin/vtracer /usr/local/bin/vtracer
COPY --from=gotools /usr/local/bin/pdf /usr/local/bin/pdf
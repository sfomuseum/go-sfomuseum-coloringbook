FROM rust:latest AS rusttools

RUN rustup target add x86_64-unknown-linux-gnu \
    && rustup target add aarch64-unknown-linux-gnu \
    && cargo install --target aarch64-unknown-linux-gnu vtracer

FROM golang:1.21-alpine as gotools

RUN mkdir /build
COPY . /build/go-sfomuseum-colouringbook

RUN apk update && apk upgrade \
    #
    && cd /build/go-sfomuseum-colouringbook \
    && go build -mod vendor -ldflags="-s -w" -o /usr/local/bin/pdf cmd/pdf/main.go \
    && cd \
    && rm -rf build
    
FROM alpine

RUN apk update && apk upgrade

COPY --from=rusttools /usr/local/cargo/bin/vtracer /usr/local/bin/vtracer
COPY --from=gotools /usr/local/bin/pdf /usr/local/bin/pdf
FROM rust:latest AS rusttools

RUN cargo install vtracer

# FROM golang:1.21-alpine as gotools


FROM alpine

RUN apk update && apk upgrade

COPY --from=rusttools /usr/local/cargo/bin/vtracer /usr/local/bin/vtracer
# syntax=docker/dockerfile:1.3.0-labs
FROM golang:1.17.6-alpine3.15 as build

WORKDIR /build
COPY ./server.go ./server.go
COPY ./go.mod ./go.mod
RUN go install  && \
    cp $GOPATH/bin/traefik-example ./server

FROM alpine:3.15

COPY --from=build /build/server ./server

COPY ./www /www

RUN chmod 555 ./server
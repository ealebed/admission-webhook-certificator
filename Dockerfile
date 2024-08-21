# syntax = docker/dockerfile:experimental
#
# ----- Go Builder Image ------
#
FROM golang:1.23-alpine AS builder

# curl git bash
RUN apk add --no-cache curl git bash make
COPY --from=golangci/golangci-lint:v1.60-alpine /usr/bin/golangci-lint /usr/bin
#
# ----- Build and Test Image -----
#
FROM builder as build

# set working directorydoc
RUN mkdir -p /go/src/certificator
WORKDIR /go/src/certificator

# load dependency
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/mod go mod download

# copy sources
COPY . .

# build
RUN make
#
# ------ get latest CA certificates
#
FROM alpine:3.20 as certs
RUN apk --update add ca-certificates
# this is for debug only Alpine image
COPY --from=build /go/src/certificator/.bin/github.com/ealebed/admission-webhook-certificator /certificator
CMD ["/certificator"]
#
# ------ certificator release Docker image ------
#
FROM scratch

# copy CA certificates
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# this is the last commabd since it's never cached
COPY --from=build /go/src/certificator/.bin/github.com/ealebed/admission-webhook-certificator /certificator

ENTRYPOINT ["/certificator"]

#############      builder       #############
FROM golang:1.13.9 AS builder

ARG TARGETS=dev

WORKDIR /go/src/github.com/mandelsoft/kipxe
COPY . .

RUN make $TARGETS

############# base
FROM alpine:3.11.3 AS base

#############      kipxe     #############
FROM base AS kipxe

WORKDIR /
COPY --from=builder /go/bin/kipxe /kipxe

ENTRYPOINT ["/kipxe"]

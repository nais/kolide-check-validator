# Create a base layer with linkerd-await from a recent release.
FROM docker.io/curlimages/curl:latest as linkerd
ARG LINKERD_AWAIT_VERSION=v0.2.3
RUN curl -sSLo /tmp/linkerd-await https://github.com/linkerd/linkerd-await/releases/download/release%2F${LINKERD_AWAIT_VERSION}/linkerd-await-${LINKERD_AWAIT_VERSION}-amd64 && \
    chmod 755 /tmp/linkerd-await

# build app
FROM golang:1.22-alpine as builder
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
COPY . /src
WORKDIR /src
RUN go test -v ./...
RUN go build -a -o ./bin/kolide-check-validator ./cmd/kolide-check-validator

# package in runtime image
FROM alpine:3.13
WORKDIR /app
COPY --from=linkerd /tmp/linkerd-await /linkerd-await
COPY --from=builder /src/bin/kolide-check-validator /app/kolide-check-validator
ENTRYPOINT ["/linkerd-await", "--shutdown", "--"]
CMD  ["/app/kolide-check-validator"]

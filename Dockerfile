ARG GO_VERSION=""
FROM golang:${GO_VERSION}alpine AS builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -o bin/kolide-check-validator ./cmd/kolide-check-validator

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /src/bin/kolide-check-validator /app/kolide-check-validator
ENTRYPOINT ["/app/kolide-check-validator"]
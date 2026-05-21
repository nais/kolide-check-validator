ARG GO_VERSION="1.26"
FROM golang:${GO_VERSION} AS builder
WORKDIR /src
COPY go.* /src/
RUN go mod download
COPY . /src
RUN go build -o ./validator ./main.go

FROM gcr.io/distroless/base
COPY --from=builder /src/validator /app/validator
ENTRYPOINT ["/app/validator"]

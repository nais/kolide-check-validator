FROM golang:1.16-alpine as builder
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GO111MODULE=on
COPY . /src
WORKDIR /src
RUN go test -v ./...
RUN go build -a -o ./bin/kolide-check-validator ./cmd/kolide-check-validator

FROM alpine:3.13
WORKDIR /app
COPY --from=builder /src/bin/kolide-check-validator /app/kolide-check-validator
CMD ["/app/kolide-check-validator"]
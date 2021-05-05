all: format test build

format:
	gofmt -s -w .

test:
	go test -v ./...

build:
	go build -o ./bin/kolide-check-validator ./cmd/kolide-check-validator/
	chmod +x ./bin/kolide-check-validator
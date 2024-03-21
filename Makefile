all: format test build

format:
	gofmt -s -w .

test:
	go test -v ./...

build:
	go build -o ./bin/kolide-check-validator ./cmd/kolide-check-validator/
	chmod +x ./bin/kolide-check-validator

check: staticcheck vulncheck deadcode

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

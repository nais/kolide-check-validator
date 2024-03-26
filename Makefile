all: fmt test build check

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

test:
	go test ./...

build:
	go build -o ./bin/kolide-check-validator ./cmd/kolide-check-validator/

check: staticcheck vulncheck deadcode

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...


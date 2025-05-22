.PHONY: all
all: fmt test check build

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w ./

.PHONY: test
test:
	go test --race -v  ./...

.PHONY: check
check: staticcheck vulncheck deadcode gosec

.PHONY: staticcheck
staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: vulncheck
vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: deadcode
deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

.PHONY: gosec
gosec:
	go tool github.com/securego/gosec/v2/cmd/gosec -terse ./...

.PHONY: build
build:
	go build -o ./bin/validator ./main.go


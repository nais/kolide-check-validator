package main

import (
	"context"
	"os"

	"github.com/navikt/kolide-check-validator/internal/cmd/kolide-check-validator"
)

func main() {
	ctx := context.Background()
	exitCode := kolide_check_validator.Run(ctx)
	os.Exit(int(exitCode))
}

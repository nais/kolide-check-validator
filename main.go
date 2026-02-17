package main

import (
	"context"
	"os"

	kolidecheckvalidator "github.com/nais/kolide-check-validator/internal/cmd/kolide-check-validator"
)

func main() {
	os.Exit(int(kolidecheckvalidator.Run(context.Background())))
}

package main

import (
	"context"
	"os"

	"github.com/nais/kolide-check-validator/internal/cmd/kolide-check-validator"
)

func main() {
	os.Exit(int(kolide_check_validator.Run(context.Background())))
}

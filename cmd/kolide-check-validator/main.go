package main

import (
	"context"
	"os"

	"github.com/navikt/kolide-check-validator/internal/cmd/kolide-check-validator"
	"github.com/sirupsen/logrus"
)

const (
	exitSuccess = iota
	exitRunError
)

func main() {
	ctx := context.Background()
	log := logrus.StandardLogger()

	if err := kolide_check_validator.Run(ctx, log); err != nil {
		log.WithError(err).Errorf("error in run: %v", err)
		os.Exit(exitRunError)
	}

	os.Exit(exitSuccess)
}

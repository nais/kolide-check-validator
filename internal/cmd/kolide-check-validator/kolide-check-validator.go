package kolidecheckvalidator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/nais/kolide-check-validator/internal/kolide"
	"github.com/nais/kolide-check-validator/internal/logger"
	"github.com/nais/kolide-check-validator/internal/slack"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

type ExitCode int

const (
	exitCodeSuccess ExitCode = iota
	exitCodeEnvFileError
	exitCodeConfigError
	exitCodeLoggerError
	exitCodeRunError
)

func Run(ctx context.Context) ExitCode {
	log := logrus.StandardLogger()

	if err := loadEnvFile(log); err != nil {
		log.WithError(err).Error("error loading .env file")
		return exitCodeEnvFileError
	}

	cfg, err := newConfig(ctx, envconfig.OsLookuper())
	if err != nil {
		log.WithError(err).Error("error when processing configuration")
		return exitCodeConfigError
	}

	appLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.WithError(err).Error("error when creating application logger")
		return exitCodeLoggerError
	}

	if err = run(ctx, cfg, appLogger); err != nil {
		appLogger.WithError(err).Errorf("error in run()")
		return exitCodeRunError
	}

	return exitCodeSuccess
}

func run(ctx context.Context, cfg *Config, log logrus.FieldLogger) error {
	incompleteChecks, err := kolide.New(
		cfg.KolideAPIToken,
		log.WithField("client", "Kolide"),
		kolide.WithHTTPClient(getHTTPClient()),
	).GetIncompleteChecks(ctx)
	if err != nil {
		return fmt.Errorf("get Kolide checks: %w", err)
	}

	if len(incompleteChecks) == 0 {
		log.Info("all Kolide checks are valid, no notification sent")
		return nil
	}

	err = slack.
		New(
			cfg.SlackWebhook,
			log.WithField("client", "Slack"),
			slack.WithHTTPClient(getHTTPClient()),
		).
		Notify(ctx, incompleteChecks)
	if err != nil {
		return fmt.Errorf("notify Slack: %w", err)
	}

	return nil
}

func getHTTPClient() *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.RetryMax = 10
	return retryableClient.StandardClient()
}

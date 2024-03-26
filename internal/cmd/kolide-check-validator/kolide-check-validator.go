package kolide_check_validator

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	kac "github.com/navikt/kolide-check-validator/internal/kolide-api-client"
	"github.com/navikt/kolide-check-validator/internal/logger"
	sc "github.com/navikt/kolide-check-validator/internal/slack-client"
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
		log.WithError(err).Errorf("error loading .env file")
		return exitCodeEnvFileError
	}

	cfg, err := newConfig(ctx, envconfig.OsLookuper())
	if err != nil {
		log.WithError(err).Errorf("error when processing configuration")
		return exitCodeConfigError
	}

	appLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.WithError(err).Errorf("error when creating application logger")
		return exitCodeLoggerError
	}

	if err = run(ctx, cfg, appLogger); err != nil {
		appLogger.WithError(err).Errorf("error in run()")
		return exitCodeRunError
	}

	return exitCodeSuccess
}

func run(ctx context.Context, cfg *Config, log logrus.FieldLogger) error {
	incompleteChecks, err := kac.New(
		cfg.KolideApiToken,
		log.WithField("client", "Kolide"),
		kac.WithHttpClient(getHttpClient()),
	).GetIncompleteChecks(ctx)
	if err != nil {
		return fmt.Errorf("get Kolide checks: %w", err)
	}

	if len(incompleteChecks) == 0 {
		log.Infof("all Kolide checks are valid")
		return nil
	}

	slackClient := sc.New(cfg.SlackWebhook, log.WithField("client", "Slack"), sc.WithHttpClient(getHttpClient()))
	timeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	err = slackClient.Notify(timeout, incompleteChecks)
	cancel()

	if err != nil {
		return fmt.Errorf("notify Slack: %w", err)
	}

	return nil
}

func getHttpClient() *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.RetryMax = 10
	return retryableClient.StandardClient()
}

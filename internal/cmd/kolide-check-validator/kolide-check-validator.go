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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	kolideApiClient := kac.New(getHttpClient(log), cfg.KolideApiToken, log.WithField("client", "Kolide"))
	slackClient := sc.New(getHttpClient(log), cfg.SlackWebhook, log.WithField("client", "Slack"))

	log.Infof("validate Kolide checks")
	timeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	checks, err := kolideApiClient.GetChecks(timeout)
	cancel()

	if err != nil {
		return fmt.Errorf("get Kolide checks: %w", err)
	}

	incompleteChecks := make([]kac.Check, 0)
	for _, check := range checks {
		if !check.HasSeverityTag() {
			incompleteChecks = append(incompleteChecks, check)
		}
	}

	log.
		WithField("num_checks", len(checks)).
		WithField("num_incomplete_checks", len(incompleteChecks)).
		Infof("validated Kolide checks")

	if len(incompleteChecks) > 0 {
		timeout, cancel = context.WithTimeout(ctx, 1*time.Minute)
		err = slackClient.Notify(timeout, incompleteChecks)
		cancel()

		if err != nil {
			return fmt.Errorf("notify Slack: %w", err)
		}
	}

	return nil
}

func getHttpClient(log logrus.FieldLogger) *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = log
	retryableClient.RetryMax = 10
	return retryableClient.StandardClient()
}

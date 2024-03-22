package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	kac "github.com/navikt/kolide-check-validator/pkg/kolide-api-client"
	sc "github.com/navikt/kolide-check-validator/pkg/slack-client"
	"github.com/sirupsen/logrus"
)

const (
	exitSuccess = iota
	exitRunError
)

func main() {
	ctx := context.Background()

	log := logrus.StandardLogger()
	if err := run(ctx, log); err != nil {
		log.WithError(err).Errorf("error in run: %v", err)
		os.Exit(exitRunError)
	}

	os.Exit(exitSuccess)
}

func run(ctx context.Context, log logrus.FieldLogger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	kolideApiToken := os.Getenv("KOLIDE_API_TOKEN")
	slackWebhook := os.Getenv("SLACK_WEBHOOK")

	kolideApiClient := kac.New(getHttpClient(log), kolideApiToken, log.WithField("client", "Kolide"))
	slackClient := sc.New(getHttpClient(log), slackWebhook, log.WithField("client", "Slack"))

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

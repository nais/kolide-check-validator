package main

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	sc "github.com/nais/kolide-check-validator/pkg/slack-client"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	MaxHttpRetries = 10
)

func getHttpClient() *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.RetryMax = MaxHttpRetries

	return retryableClient.StandardClient()
}

func main() {
	kolideApiToken := os.Getenv("KOLIDE_API_TOKEN")
	slackWebhook := os.Getenv("SLACK_WEBHOOK")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kolideApiClient := kac.New(getHttpClient(), kolideApiToken)
	slackClient := sc.New(getHttpClient(), slackWebhook)

	log.Infof("validate Kolide checks")
	var incompleteChecks []kac.Check

	timeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	checks, err := kolideApiClient.GetChecks(timeout)
	cancel()

	if err != nil {
		log.Fatalf("get checks: %v", err)
	}

	for _, check := range checks {
		if !hasSeverityTag(check) {
			incompleteChecks = append(incompleteChecks, check)
		}
	}

	log.Infof("found %d checks (%d incomplete)", len(checks), len(incompleteChecks))

	if len(incompleteChecks) > 0 {
		timeout, cancel = context.WithTimeout(ctx, 1*time.Minute)
		err = slackClient.Notify(timeout, incompleteChecks)
		cancel()

		if err != nil {
			log.Fatalf("notify Slack: %v", err)
		}
	}
}

func hasSeverityTag(check kac.Check) bool {
	severityTags := []string{"info", "notice", "warning", "danger", "critical"}
	for _, tag := range check.Tags {
		tag = strings.ToLower(tag)
		for _, severityTag := range severityTags {
			if tag == severityTag {
				return true
			}
		}
	}

	return false
}

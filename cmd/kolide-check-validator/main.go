package main

import (
	"context"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	sc "github.com/nais/kolide-check-validator/pkg/slack-client"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

const CronInterval = 5 * time.Minute

var (
	kolideApiToken string
	slackWebhook   string
)

func init() {
	kolideApiToken = os.Getenv("KOLIDE_API_TOKEN")
	slackWebhook = os.Getenv("SLACK_WEBHOOK")
}

func main() {
	mainContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	kolideApiClient := kac.New(kolideApiToken)
	slackClient := sc.New(slackWebhook)
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
			ticker.Reset(CronInterval)

			var incompleteChecks []kac.Check
			timeout, cancel := context.WithTimeout(mainContext, 1*time.Minute)

			log.Infof("Validate all Kolide checks for severity tag(s)")
			checks, err := kolideApiClient.GetChecks(timeout)
			cancel()

			if err != nil {
				log.Errorf("get checks: %v", err)
				continue
			}

			log.Infof("Fetched %d checks", len(checks))

			for _, check := range checks {
				if !hasSeverityTag(check) {
					incompleteChecks = append(incompleteChecks, check)
				}
			}

			log.Infof("Found %d incomplete check(s)", len(incompleteChecks))

			if len(incompleteChecks) > 0 {
				timeout, cancel := context.WithTimeout(mainContext, 1*time.Minute)
				log.Infof("Send alert to Slack")
				slackClient.Notify(timeout, incompleteChecks)
				cancel()
			}

		case <-mainContext.Done():
			return
		}
	}
}

func hasSeverityTag(check kac.Check) bool {
	for _, tag := range check.Tags {
		switch strings.ToLower(tag) {
		case
			"info",
			"notice",
			"warning",
			"danger",
			"critical":
			return true
		}
	}

	return false
}

package main

import (
	"context"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

const CronInterval = 5 * time.Minute

var kolideApiToken string

func init() {
	kolideApiToken = os.Getenv("KOLIDE_API_TOKEN")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kolideApiClient := kac.New(kolideApiToken)
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
			ticker.Reset(CronInterval)
			var incompleteChecks []kac.Check

			log.Infof("Validate all Kolide checks for severity tag(s)")
			checks, err := kolideApiClient.GetChecks()

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
				// alert Slack
			}

		case <-ctx.Done():
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

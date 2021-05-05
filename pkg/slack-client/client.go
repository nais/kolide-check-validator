package slack_client

import (
	"context"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	"net/http"
)

// New returns a new SlackClient instance
func New(slackWebhook string) *SlackClient {
	return &SlackClient{
		slackWebhook: slackWebhook,
		client:       &http.Client{},
	}
}

// Notify will send a message to a Slack channel via `SlackClient.slackWebhook`
func (sc *SlackClient) Notify(context context.Context, checks []kac.Check) error {
	return nil
}

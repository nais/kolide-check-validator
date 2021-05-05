package slack_client

import "net/http"

type SlackClient struct {
	slackWebhook string
	client       *http.Client
}

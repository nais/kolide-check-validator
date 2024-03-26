package slack_client

import "net/http"

type Option func(client *SlackClient)

func WithHttpClient(client *http.Client) Option {
	return func(c *SlackClient) {
		c.client = client
	}
}

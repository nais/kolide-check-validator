package kolide_api_client

import "net/http"

type Option func(*KolideClient)

func WithBaseUrl(baseUrl string) Option {
	return func(c *KolideClient) {
		c.baseUrl = baseUrl
	}
}

func WithHttpClient(client *http.Client) Option {
	return func(c *KolideClient) {
		c.client.client = client
	}
}

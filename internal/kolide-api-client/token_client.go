package kolide_api_client

import "net/http"

type httpClient struct {
	client   *http.Client
	apiToken string
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

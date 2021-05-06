package kolide_api_client

import (
	"fmt"
	"net/http"
)

type Transport struct {
	apiToken string
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultTransport.RoundTrip(req)
}

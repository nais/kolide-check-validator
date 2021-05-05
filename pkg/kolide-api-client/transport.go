package kolide_api_client

import (
	"fmt"
	"net/http"
)

type Transport struct {
	apiToken string
}

// RoundTrip sets the `Authorization` and ´Content-Type` headers before issuing the request
// The resulting response and an possible error is returned
func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultTransport.RoundTrip(req)
}

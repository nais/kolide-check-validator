package kolide_api_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
)

func New(apiToken string, log logrus.FieldLogger, opts ...Option) *KolideClient {
	c := &KolideClient{
		baseUrl: "https://k2.kolide.com/api/v0",
		client: &httpClient{
			client:   http.DefaultClient,
			apiToken: apiToken,
		},
		log: log,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (kc *KolideClient) GetChecks(ctx context.Context) ([]Check, error) {
	apiUrl, err := url.Parse(kc.baseUrl + "/checks")
	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiUrl.Query()
	query.Set("per_page", strconv.Itoa(100))
	apiUrl.RawQuery = query.Encode()

	checks := make([]Check, 0)
	for {
		paginatedChecks, nextCursor, err := kc.getPaginatedChecks(ctx, apiUrl)
		if err != nil {
			return nil, err
		}

		checks = append(checks, paginatedChecks...)

		if nextCursor == "" {
			break
		}

		query.Set("cursor", nextCursor)
		apiUrl.RawQuery = query.Encode()
	}

	return checks, nil
}

func (kc *KolideClient) getPaginatedChecks(ctx context.Context, url *url.URL) ([]Check, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := kc.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("get paginated response: %w", err)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d (%s)", resp.StatusCode, string(bytes))
	}

	checksResponse := ChecksResponse{}
	if err = json.Unmarshal(bytes, &checksResponse); err != nil {
		return nil, "", fmt.Errorf("decode paginated response: %w", err)
	}

	return checksResponse.Checks, checksResponse.Pagination.NextCursor, nil
}

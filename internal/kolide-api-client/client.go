package kolide_api_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (c *KolideClient) GetIncompleteChecks(ctx context.Context) ([]Check, error) {
	apiUrl, err := url.Parse(c.baseUrl + "/checks")
	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiUrl.Query()
	query.Set("per_page", strconv.Itoa(10))
	apiUrl.RawQuery = query.Encode()

	numChecks := 0
	incompleteChecks := make([]Check, 0)
	for {
		paginatedChecks, nextCursor, err := c.getPaginatedChecks(ctx, apiUrl)
		if err != nil {
			return nil, err
		}

		for _, check := range paginatedChecks {
			numChecks++
			if !check.HasSeverityTag() {
				incompleteChecks = append(incompleteChecks, check)
			}
		}

		if nextCursor == "" {
			break
		}

		query.Set("cursor", nextCursor)
		apiUrl.RawQuery = query.Encode()
	}

	c.log.
		WithField("num_checks", numChecks).
		WithField("num_incomplete_checks", len(incompleteChecks)).
		Infof("validated Kolide checks")

	return incompleteChecks, nil
}

func (c *KolideClient) getPaginatedChecks(ctx context.Context, url *url.URL) ([]Check, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
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

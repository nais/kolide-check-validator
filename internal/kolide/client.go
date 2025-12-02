package kolide

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type OptionFunc func(*Client)

func WithBaseUrl(baseUrl string) OptionFunc {
	return func(c *Client) { c.baseUrl = baseUrl }
}

func WithHttpClient(client *http.Client) OptionFunc {
	return func(c *Client) { c.httpClient.client = client }
}

type Client struct {
	baseUrl    string
	httpClient *httpClient
	log        logrus.FieldLogger
}

type ChecksResponse struct {
	Checks     []Check `json:"data"`
	Pagination struct {
		NextCursor string `json:"next_cursor"`
	} `json:"pagination"`
}

type CheckTag struct {
	Name string `json:"name"`
}

type Check struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Tags []CheckTag `json:"check_tags"`
}

func (c *Check) HasSeverityTag() bool {
	if c == nil {
		return false
	}

	severityTags := []string{"info", "notice", "warning", "danger", "critical"}
	for _, tag := range c.Tags {
		tagName := strings.ToLower(tag.Name)
		for _, severityTag := range severityTags {
			if tagName == severityTag {
				return true
			}
		}
	}

	return false
}

type httpClient struct {
	client   *http.Client
	apiToken string
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Kolide-API-Version", "2023-05-26")

	return c.client.Do(req)
}

func New(apiToken string, log logrus.FieldLogger, opts ...OptionFunc) *Client {
	c := &Client{
		baseUrl: "https://api.kolide.com",
		httpClient: &httpClient{
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

func (c *Client) GetIncompleteChecks(ctx context.Context) ([]Check, error) {
	apiUrl, err := url.Parse(c.baseUrl + "/checks")
	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiUrl.Query()
	query.Set("per_page", strconv.Itoa(100))
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
		Info("validated Kolide checks")

	return incompleteChecks, nil
}

func (c *Client) getPaginatedChecks(ctx context.Context, url *url.URL) ([]Check, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
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

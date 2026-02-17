package kolide

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type OptionFunc func(*Client)

func WithBaseURL(baseURL string) OptionFunc {
	return func(c *Client) { c.baseURL = baseURL }
}

func WithHTTPClient(client *http.Client) OptionFunc {
	return func(c *Client) { c.httpClient.client = client }
}

type Client struct {
	baseURL    string
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

	severityTags := []string{"info", "notice", "attention", "warning", "danger", "critical"}
	for _, tag := range c.Tags {
		tagName := strings.ToLower(tag.Name)
		if slices.Contains(severityTags, tagName) {
			return true
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

	return c.client.Do(req) // #nosec G704 -- URL is validated by validateBaseURL
}

func New(apiToken string, log logrus.FieldLogger, opts ...OptionFunc) *Client {
	c := &Client{
		baseURL: "https://api.kolide.com",
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
	if err := validateBaseURL(c.baseURL); err != nil {
		return nil, fmt.Errorf("invalid Kolide base URL: %w", err)
	}

	apiURL, err := url.Parse(c.baseURL + "/checks")
	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiURL.Query()
	query.Set("per_page", strconv.Itoa(100))
	apiURL.RawQuery = query.Encode()

	numChecks := 0
	incompleteChecks := make([]Check, 0)
	for {
		paginatedChecks, nextCursor, err := c.getPaginatedChecks(ctx, apiURL)
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
		apiURL.RawQuery = query.Encode()
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
	defer resp.Body.Close()

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

func validateBaseURL(rawBaseURL string) error {
	parsed, err := url.Parse(rawBaseURL)
	if err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}

	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname())) {
		return fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	return ip.IsLoopback()
}

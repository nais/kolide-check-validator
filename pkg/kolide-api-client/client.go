package kolide_api_client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	ApiBaseUrl           = "https://k2.kolide.com/api/v0"
	ApiResultsPerRequest = 100
)

func New(client *http.Client, apiToken string) *KolideClient {
	return NewConfiguredClient(client, ApiBaseUrl, apiToken)
}

func NewConfiguredClient(client *http.Client, baseUrl, apiToken string) *KolideClient {
	kolideApiTransport := &Transport{
		apiToken:        apiToken,
		parentTransport: client.Transport,
	}

	return &KolideClient{
		baseUrl: baseUrl,
		client: &http.Client{
			Transport: kolideApiTransport,
		},
	}
}

func (kc *KolideClient) GetChecks(ctx context.Context) ([]Check, error) {
	var checks []Check

	apiUrl, err := url.Parse(fmt.Sprintf("%s/checks", kc.baseUrl))
	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiUrl.Query()
	query.Set("per_page", strconv.Itoa(ApiResultsPerRequest))
	apiUrl.RawQuery = query.Encode()

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

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d (%s)", resp.StatusCode, string(bytes))
	}

	var checksResponse ChecksResponse

	err = json.Unmarshal(bytes, &checksResponse)
	if err != nil {
		return nil, "", fmt.Errorf("decode paginated response: %w", err)
	}

	return checksResponse.Checks, checksResponse.Pagination.NextCursor, nil
}

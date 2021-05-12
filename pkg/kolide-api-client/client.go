package kolide_api_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	ApiBaseUrl           = "https://k2.kolide.com/api/v0"
	ApiResultsPerRequest = 100
	MaxHttpRetries       = 10
)

func New(apiToken string) *KolideClient {
	client := retryablehttp.NewClient()
	client.HTTPClient = &http.Client{Transport: Transport{
		apiToken: apiToken,
	}}
	client.Logger = nil
	client.RetryMax = MaxHttpRetries

	return &KolideClient{
		baseUrl: ApiBaseUrl,
		client:  client,
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
	req, err := retryablehttp.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	resp, err := kc.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, "", fmt.Errorf("get paginated response: %w", err)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response body: %w", err)
	}

	var checksResponse ChecksResponse

	err = json.Unmarshal(bytes, &checksResponse)
	if err != nil {
		return nil, "", fmt.Errorf("decode paginated response: %w", err)
	}

	return checksResponse.Checks, checksResponse.Pagination.NextCursor, nil
}

package kolide_api_client

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const ApiBaseUrl = "https://k2.kolide.com/api/v0"
const ApiResultsPerRequest = 100
const DefaultRetryAfter = 5 * time.Second
const MaxHttpRetries = 10

// New returns a new KolideClient instance
func New(apiToken string) *KolideClient {
	return &KolideClient{
		baseUrl: ApiBaseUrl,
		client: &http.Client{Transport: Transport{
			apiToken: apiToken,
		}},
	}
}

// get will issue a request for the given path, and provide a response or an error
// Each request will be tried multiple times if a failure occurs, and rate limiting
// will be upheld if the response includes a Retry-After header.
func (kc *KolideClient) get(path string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, path, nil)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for attempt := 0; attempt < MaxHttpRetries; attempt++ {
		response, err := kc.client.Do(request)

		if err != nil {
			return nil, err
		}

		switch statusCode := response.StatusCode; {
		case statusCode == http.StatusOK:
			return response, nil
		case statusCode == http.StatusTooManyRequests:
			backoff := getRetryAfter(response.Header)
			log.Infof("Rate limited by Kolide, wait %d second(s)", backoff)
			wait(backoff)
		case statusCode >= 500:
			backoff := int(math.Pow(float64(attempt+1), 2))
			log.Infof("Internal server error, wait %d second(s)", backoff)
			wait(time.Duration(backoff) * time.Second)
		default:
			return nil, fmt.Errorf("unexpected stauts code: %d, response: %v", statusCode, response)
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// GetChecks will return all checks form the Kolide API
func (kc *KolideClient) GetChecks() ([]Check, error) {
	var checks []Check
	cursor := ""

	apiUrl, err := url.Parse(fmt.Sprintf("%s/checks", kc.baseUrl))

	if err != nil {
		return nil, fmt.Errorf("create URL: %w", err)
	}

	query := apiUrl.Query()
	query.Set("per_page", strconv.Itoa(ApiResultsPerRequest))
	apiUrl.RawQuery = query.Encode()

	for {
		response, err := kc.get(apiUrl.String())

		if err != nil {
			return nil, fmt.Errorf("get paginated response: %w", err)
		}

		responseBytes, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}

		var checksResponse ChecksResponse
		err = json.Unmarshal(responseBytes, &checksResponse)

		if err != nil {
			return nil, fmt.Errorf("decode paginated response: %w", err)
		}

		checks = append(checks, checksResponse.Checks...)

		cursor = checksResponse.Pagination.NextCursor

		if cursor == "" {
			break
		}

		query.Set("cursor", cursor)
		apiUrl.RawQuery = query.Encode()
	}

	return checks, nil
}

// wait sleeps for `sleep` + 0..10 seconds
func wait(sleep time.Duration) {
	time.Sleep(sleep + (time.Second * time.Duration(rand.Intn(10))))
}

// getRetryAfter parses the Retry-After header and returns a duration in seconds
// If the header is missing or is not understood, the default retry after value will be returned, defined by the
// DefaultRetryAfter constant
func getRetryAfter(headers http.Header) time.Duration {
	retryAfter := headers.Get("Retry-After")

	if retryAfter == "" {
		return 0
	}

	seconds, err := strconv.Atoi(retryAfter)

	if err != nil {
		retryAfterDate, err := time.Parse(time.RFC1123, retryAfter)

		if err != nil || retryAfterDate.Before(time.Now()) {
			return DefaultRetryAfter
		}

		return time.Until(retryAfterDate).Round(time.Second)
	}

	if seconds < 0 {
		return DefaultRetryAfter
	}

	return time.Second * time.Duration(seconds)
}

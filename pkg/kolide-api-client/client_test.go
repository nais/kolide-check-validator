package kolide_api_client_test

import (
	"context"
	"fmt"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	apiToken = "test token"
)

func getTestServer(t *testing.T, pages map[string]string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/checks", func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, fmt.Sprintf("Bearer %s", apiToken), request.Header.Get("Authorization"))

		cursor := request.URL.Query().Get("cursor")

		_, err := fmt.Fprintf(writer, pages[cursor])
		assert.NoError(t, err)
	})
	return httptest.NewServer(mux)
}

func TestKolideClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	t.Run("no checks", func(t *testing.T) {
		pages := map[string]string{
			"": `{"data":[],"pagination":{"next_cursor":""}}`,
		}

		testServer := getTestServer(t, pages)

		apiClient := kac.NewConfiguredClient(testServer.Client(), testServer.URL, apiToken)

		checks, err := apiClient.GetChecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, checks, 0)
	})

	t.Run("invalid response body", func(t *testing.T) {
		pages := map[string]string{
			"": `some string`,
		}

		testServer := getTestServer(t, pages)

		apiClient := kac.NewConfiguredClient(testServer.Client(), testServer.URL, apiToken)

		checks, err := apiClient.GetChecks(ctx)

		assert.Nil(t, checks)
		assert.Error(t, err)
	})

	t.Run("non 200 OK", func(t *testing.T) {
		mux := http.NewServeMux()
		testServer := httptest.NewServer(mux)

		apiClient := kac.NewConfiguredClient(testServer.Client(), testServer.URL, apiToken)

		checks, err := apiClient.GetChecks(ctx)

		assert.Nil(t, checks)
		assert.Error(t, err)
	})

	t.Run("multiple pages of checks", func(t *testing.T) {
		pages := map[string]string{
			"":      `{"data":[{"id":1},{"id":2}],"pagination":{"next_cursor":"page2"}}`,
			"page2": `{"data":[{"id":3}],"pagination": {"next_cursor":""}}`,
		}

		testServer := getTestServer(t, pages)

		apiClient := kac.NewConfiguredClient(testServer.Client(), testServer.URL, apiToken)

		checks, err := apiClient.GetChecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, checks, 3)
		assert.Equal(t, 1, checks[0].Id)
		assert.Equal(t, 2, checks[1].Id)
		assert.Equal(t, 3, checks[2].Id)
	})
}

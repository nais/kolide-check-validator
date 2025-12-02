package kolide_api_client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	kac "github.com/nais/kolide-check-validator/internal/kolide-api-client"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var apiToken = "test token"

func getTestServer(t *testing.T, pages map[string]string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/checks", func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "Bearer "+apiToken, request.Header.Get("Authorization"))

		cursor := request.URL.Query().Get("cursor")

		_, err := fmt.Fprint(writer, pages[cursor])
		assert.NoError(t, err)
	})
	return httptest.NewServer(mux)
}

func getKolideClientForTestServer(server *httptest.Server, log logrus.FieldLogger) *kac.KolideClient {
	return kac.New(apiToken, log, kac.WithHttpClient(server.Client()), kac.WithBaseUrl(server.URL))
}

func TestKolideClient(t *testing.T) {
	log, _ := logrustest.NewNullLogger()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	t.Run("no checks", func(t *testing.T) {
		pages := map[string]string{
			"": `{"data":[],"pagination":{"next_cursor":""}}`,
		}

		testServer := getTestServer(t, pages)

		apiClient := getKolideClientForTestServer(testServer, log)

		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, incompleteChecks, 0)
	})

	t.Run("invalid response body", func(t *testing.T) {
		pages := map[string]string{
			"": `some string`,
		}

		testServer := getTestServer(t, pages)

		apiClient := getKolideClientForTestServer(testServer, log)

		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		assert.Nil(t, incompleteChecks)
		assert.Error(t, err)
	})

	t.Run("non 200 OK", func(t *testing.T) {
		mux := http.NewServeMux()
		testServer := httptest.NewServer(mux)

		apiClient := getKolideClientForTestServer(testServer, log)

		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		assert.Nil(t, incompleteChecks)
		assert.Error(t, err)
	})

	t.Run("multiple pages of checks", func(t *testing.T) {
		pages := map[string]string{
			"":      `{"data":[{"id":1},{"id":2,"tags":["foo"]}],"pagination":{"next_cursor":"page2"}}`,
			"page2": `{"data":[{"id":3,"tags":["info"]},{"id":4}],"pagination": {"next_cursor":"page3"}}`,
			"page3": `{"data":[{"id":5,"tags":["notice"]}],"pagination": {"next_cursor":""}}`,
		}

		testServer := getTestServer(t, pages)

		apiClient := getKolideClientForTestServer(testServer, log)

		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		assert.NoError(t, err)
		assert.Len(t, incompleteChecks, 3)
		assert.Equal(t, 1, incompleteChecks[0].Id)
		assert.Equal(t, 2, incompleteChecks[1].Id)
		assert.Equal(t, 4, incompleteChecks[2].Id)
	})
}

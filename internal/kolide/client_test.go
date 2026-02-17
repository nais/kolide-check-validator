package kolide_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nais/kolide-check-validator/internal/kolide"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var apiToken = "test token"

func TestKolideClient(t *testing.T) {
	log, _ := logrustest.NewNullLogger()
	ctx := context.Background()

	t.Run("no checks", func(t *testing.T) {
		pages := map[string]string{
			"": `{"data":[],"pagination":{"next_cursor":""}}`,
		}
		testServer := getTestServer(t, pages)
		apiClient := getKolideClientForTestServer(testServer, log)
		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(incompleteChecks) != 0 {
			t.Fatalf("expected 0 incomplete checks, got %d", len(incompleteChecks))
		}
	})

	t.Run("invalid response body", func(t *testing.T) {
		pages := map[string]string{
			"": `some string`,
		}
		testServer := getTestServer(t, pages)
		apiClient := getKolideClientForTestServer(testServer, log)
		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if contains := "invalid character"; !strings.Contains(err.Error(), contains) {
			t.Fatalf("expected error to contain %q, got: %s", contains, err.Error())
		}

		if len(incompleteChecks) != 0 {
			t.Fatalf("expected 0 incomplete checks, got %d", len(incompleteChecks))
		}
	})

	t.Run("non 200 OK", func(t *testing.T) {
		mux := http.NewServeMux()
		testServer := httptest.NewServer(mux)
		apiClient := getKolideClientForTestServer(testServer, log)
		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if contains := "unexpected status code: 404"; !strings.Contains(err.Error(), contains) {
			t.Fatalf("expected error to contain %q, got: %s", contains, err.Error())
		}

		if len(incompleteChecks) != 0 {
			t.Fatalf("expected 0 incomplete checks, got %d", len(incompleteChecks))
		}
	})

	t.Run("multiple pages of checks", func(t *testing.T) {
		pages := map[string]string{
			"":      `{"data":[{"id":"1"},{"id":"2","check_tags":[{"name":"foo"}]}],"pagination":{"next_cursor":"page2"}}`,
			"page2": `{"data":[{"id":"3","check_tags":[{"name":"info"}]},{"id":"4"}],"pagination": {"next_cursor":"page3"}}`,
			"page3": `{"data":[{"id":"5","check_tags":[{"name":"notice"}]}],"pagination": {"next_cursor":""}}`,
		}
		testServer := getTestServer(t, pages)
		apiClient := getKolideClientForTestServer(testServer, log)
		incompleteChecks, err := apiClient.GetIncompleteChecks(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(incompleteChecks) != 3 {
			t.Fatalf("expected 3 incomplete checks, got %d", len(incompleteChecks))
		}

		if incompleteChecks[0].ID != "1" {
			t.Fatalf(`expected first incomplete check ID to be "1", got %q`, incompleteChecks[0].ID)
		}

		if incompleteChecks[1].ID != "2" {
			t.Fatalf(`expected second incomplete check ID to be "2", got %q`, incompleteChecks[1].ID)
		}

		if incompleteChecks[2].ID != "4" {
			t.Fatalf(`expected third incomplete check ID to be "4", got %q`, incompleteChecks[2].ID)
		}
	})
}

func TestCheck_HasSeverityTag(t *testing.T) {
	tests := []struct {
		name           string
		check          kolide.Check
		hasSeverityTag bool
	}{
		{
			name: "check with no tags",
			check: kolide.Check{
				Tags: nil,
			},
			hasSeverityTag: false,
		},
		{
			name: "check with no severity tags",
			check: kolide.Check{
				Tags: []kolide.CheckTag{{Name: "foo"}, {Name: "bar"}},
			},
			hasSeverityTag: false,
		},
		{
			name: "check with severity tag",
			check: kolide.Check{
				Tags: []kolide.CheckTag{{Name: "info"}, {Name: "foo"}, {Name: "bar"}},
			},
			hasSeverityTag: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check.HasSeverityTag(); got != tt.hasSeverityTag {
				t.Errorf("HasSeverityTag(%v) = %v, expected %v", tt.check.Tags, got, tt.hasSeverityTag)
			}
		})
	}
}

func getTestServer(t *testing.T, pages map[string]string) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/checks", func(w http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		if expected := "Bearer " + apiToken; auth != expected {
			t.Fatalf("expected Authorization header to be %q, got %q", expected, auth)
		}
		_, _ = fmt.Fprint(w, pages[req.URL.Query().Get("cursor")])
	})
	return httptest.NewServer(mux)
}

func getKolideClientForTestServer(server *httptest.Server, log logrus.FieldLogger) *kolide.Client {
	return kolide.New(apiToken, log, kolide.WithHTTPClient(server.Client()), kolide.WithBaseURL(server.URL))
}

package slack_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nais/kolide-check-validator/internal/kolide"
	"github.com/nais/kolide-check-validator/internal/slack"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestSlackClient(t *testing.T) {
	log, _ := logrustest.NewNullLogger()
	ctx := context.Background()

	t.Run("response status not 200 OK", func(t *testing.T) {
		slackClient := getSlackClientForTestServer(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(500)
		}, log)

		err := slackClient.Notify(ctx, []kolide.Check{{
			ID:   "1",
			Name: "check",
		}})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if contains := "unable to notify Slack: HTTP 500"; !strings.Contains(err.Error(), contains) {
			t.Fatalf("expected error to contain %q, got: %s", contains, err.Error())
		}
	})

	t.Run("should fail when no checks are passed", func(t *testing.T) {
		slackClient := getSlackClientForTestServer(func(http.ResponseWriter, *http.Request) {
			t.Fail() // should not be called
		}, log)

		err := slackClient.Notify(ctx, []kolide.Check{})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if contains := "no checks"; !strings.Contains(err.Error(), contains) {
			t.Fatalf("expected error to contain %q, got: %s", contains, err.Error())
		}
	})

	t.Run("can notify Slack", func(t *testing.T) {
		slackClient := getSlackClientForTestServer(func(_ http.ResponseWriter, req *http.Request) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}

			contains := []string{
				"The following Kolide checks",
				"check 1",
				"check 2",
			}

			bodyString := string(body)
			for _, c := range contains {
				if !strings.Contains(bodyString, c) {
					t.Fatalf("expected body to contain %q", c)
				}
			}
		}, log)

		err := slackClient.Notify(ctx, []kolide.Check{
			{
				ID:   "1",
				Name: "check 1",
			},
			{
				ID:   "2",
				Name: "check 2",
			},
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}

func getSlackClientForTestServer(handler func(writer http.ResponseWriter, request *http.Request), log logrus.FieldLogger) *slack.Client {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	server := httptest.NewServer(mux)
	return slack.New(server.URL, log, slack.WithHttpClient(server.Client()))
}

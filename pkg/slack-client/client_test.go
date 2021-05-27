package slack_client

import (
	"context"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_mrkdown(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "no mrkdown",
			input:  "just a string",
			output: "just a string",
		},
		{
			name:   "bold",
			input:  "some **string** *with* **bold**",
			output: "some *string* *with* *bold*",
		},
		{
			name:   "links",
			input:  "a string [with a link](someurl).",
			output: "a string <someurl|with a link>.",
		},
		{
			name:  "paragraphs",
			input: "some  new  paragraph.",
			output: `some

new

paragraph.`,
		},
		{
			name:  "all conversions",
			input: "This **is** a [link](url).  New paragraph.",
			output: `This *is* a <url|link>.

New paragraph.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mrkdown(tt.input); got != tt.output {
				t.Errorf("mrkdown(%s) = %s, want %s", tt.input, got, tt.output)
			}
		})
	}
}

func Test_s(t *testing.T) {
	tests := []struct {
		name   string
		input  int
		output string
	}{
		{
			name:   "plural",
			input:  2,
			output: "s",
		},
		{
			name:   "not plural",
			input:  1,
			output: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s(tt.input); got != tt.output {
				t.Errorf("s(%d) = %s, want %s", tt.input, got, tt.output)
			}
		})
	}
}

func Test_na(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output string
	}{
		{
			name:   "empty list",
			input:  nil,
			output: "N/A",
		},
		{
			name:   "list with entries",
			input:  []string{"foo", "bar"},
			output: "foo, bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := na(tt.input); got != tt.output {
				t.Errorf("na(%v) = %s, want %v", tt.input, got, tt.output)
			}
		})
	}
}

func getSlackClientForTestServer(handler func(writer http.ResponseWriter, request *http.Request)) *SlackClient {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	server := httptest.NewServer(mux)

	return New(server.Client(), server.URL)
}

func TestSlackClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	t.Run("response status not 200 OK", func(t *testing.T) {
		apiClient := getSlackClientForTestServer(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(500)
		})

		err := apiClient.Notify(ctx, []kac.Check{
			{
				Name:        "check",
				Description: "description",
			},
		})

		assert.Error(t, err)
	})

	t.Run("should fail when no checks are passed", func(t *testing.T) {
		apiClient := getSlackClientForTestServer(func(writer http.ResponseWriter, request *http.Request) {
			t.Fail()
		})

		err := apiClient.Notify(ctx, []kac.Check{})

		assert.Error(t, err)
	})

	t.Run("can notify Slack", func(t *testing.T) {
		apiClient := getSlackClientForTestServer(func(writer http.ResponseWriter, request *http.Request) {
			body, err := ioutil.ReadAll(request.Body)
			assert.NoError(t, err)

			bodyString := string(body)

			assert.True(t, strings.Contains(bodyString, "The following Kolide checks"))
			assert.True(t, strings.Contains(bodyString, "check 1"))
			assert.True(t, strings.Contains(bodyString, "description 1"))
			assert.True(t, strings.Contains(bodyString, "comp 1"))
			assert.True(t, strings.Contains(bodyString, "topic 1"))
			assert.True(t, strings.Contains(bodyString, "check 2"))
			assert.True(t, strings.Contains(bodyString, "description 2"))
			assert.True(t, strings.Contains(bodyString, "comp 2"))
			assert.True(t, strings.Contains(bodyString, "topic 2"))
		})

		err := apiClient.Notify(ctx, []kac.Check{
			{
				Name:          "check 1",
				Description:   "description 1",
				Compatibility: []string{"comp 1"},
				Topics:        []string{"topic 1"},
			},
			{
				Name:          "check 2",
				Description:   "description 2",
				Compatibility: []string{"comp 2"},
				Topics:        []string{"topic 2"},
			},
		})

		assert.NoError(t, err)
	})
}

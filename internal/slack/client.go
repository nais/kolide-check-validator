package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nais/kolide-check-validator/internal/kolide"
	"github.com/sirupsen/logrus"
)

type Payload struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type     string      `json:"type"`
	Text     *TextObject `json:"text,omitempty"`
	Elements []Element   `json:"elements,omitempty"`
	Style    string      `json:"style,omitempty"`
	Indent   int         `json:"indent,omitempty"`
	URL      string      `json:"url,omitempty"`
}

type TextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}

type Element struct {
	Type     string    `json:"type"`
	Style    string    `json:"style,omitempty"`
	Indent   int       `json:"indent,omitempty"`
	URL      string    `json:"url,omitempty"`
	Text     string    `json:"text,omitempty"`
	Elements []Element `json:"elements,omitempty"`
}

type OptionFunc func(client *Client)

func WithHttpClient(client *http.Client) OptionFunc {
	return func(c *Client) { c.httpClient = client }
}

type Client struct {
	slackWebhook string
	httpClient   *http.Client
	log          logrus.FieldLogger
}

func New(slackWebhook string, log logrus.FieldLogger, opts ...OptionFunc) *Client {
	c := &Client{
		slackWebhook: slackWebhook,
		httpClient:   http.DefaultClient,
		log:          log,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Notify(ctx context.Context, checks []kolide.Check) error {
	if len(checks) == 0 {
		return fmt.Errorf("no checks")
	}

	body, err := getRequestBody(checks)
	if err != nil {
		return fmt.Errorf("get request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.slackWebhook, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			responseBytes = []byte("unable to read error response body")
		}

		return fmt.Errorf("unable to notify Slack: HTTP %d: %s", resp.StatusCode, responseBytes)
	}

	c.log.Info("notification sent")
	return nil
}

func getRequestBody(checks []kolide.Check) (io.Reader, error) {
	listItems := make([]Element, len(checks))
	for i, check := range checks {
		listItems[i] = Element{
			Type: "rich_text_section",
			Elements: []Element{
				{
					Type: "link",
					URL:  "https://app.kolide.com/1401/checks/" + check.ID + "/results/all",
					Text: check.Name,
				},
			},
		}
	}

	payload := Payload{
		Blocks: []Block{
			{
				Type: "header",
				Text: &TextObject{
					Type:  "plain_text",
					Text:  ":warning: The following Kolide checks are missing severity tags: :warning:",
					Emoji: true,
				},
			},
			{
				Type: "rich_text",
				Elements: []Element{{
					Type:     "rich_text_list",
					Style:    "bullet",
					Indent:   0,
					Elements: listItems,
				}},
			},
			{
				Type: "context",
				Elements: []Element{
					{
						Type: "mrkdwn",
						Text: "This message has been brought to you by <https://github.com/nais/kolide-check-validator|nais/kolide-check-validator>",
					},
				},
			},
		},
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	return body, nil
}

package slack_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	kac "github.com/nais/kolide-check-validator/internal/kolide-api-client"
	"github.com/sirupsen/logrus"
)

func New(slackWebhook string, log logrus.FieldLogger, opts ...Option) *SlackClient {
	c := &SlackClient{
		slackWebhook: slackWebhook,
		client:       http.DefaultClient,
		log:          log,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *SlackClient) Notify(ctx context.Context, checks []kac.Check) error {
	if len(checks) == 0 {
		return fmt.Errorf("no checks")
	}

	body, err := getRequestBody(checks)
	if err != nil {
		return fmt.Errorf("get request body: %w", err)
	}

	c.log.Debugf("request payload for notification: %s", body)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.slackWebhook, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
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

	c.log.Infof("notification sent")
	return nil
}

func getRequestBody(checks []kac.Check) ([]byte, error) {
	blocks := []Block{
		{
			Type: "header",
			Text: &Text{
				Type: "plain_text",
				Text: ":warning: The following Kolide checks are missing severity tags: :warning:",
				Emoji: func() *bool {
					t := true
					return &t
				}(),
			},
		},
	}

	divider := Block{Type: "divider"}
	for _, check := range checks {
		blocks = append(blocks, divider, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: fmt.Sprintf(
					"*<https://k2.kolide.com/1401/checks/%d|%s>* - *%d failure%s*\n%s\n\nCompatibility: _%s_, Topics: _%s_, Tags: _%s_",
					check.Id,
					check.Name,
					check.FailingDeviceCount,
					S(check.FailingDeviceCount),
					Mrkdown(check.Description),
					Join(check.Compatibility),
					Join(check.Topics),
					Join(check.Tags),
				),
			},
		})
	}

	blocks = append(blocks, divider, Block{
		Type: "context",
		Elements: []Text{
			{
				Type: "mrkdwn",
				Text: "This message has been brought to you by <https://github.com/nais/kolide-check-validator|nais/kolide-check-validator>",
			},
		},
	})

	body, err := json.Marshal(&Message{
		Blocks: blocks,
	})
	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	return body, nil
}

func Mrkdown(string string) string {
	bold, _ := regexp.Compile(`\*\*`)
	string = bold.ReplaceAllString(string, "*")

	links, _ := regexp.Compile(`\[(.*?)\]\((.*?)\)`)
	string = links.ReplaceAllString(string, "<$2|$1>")

	paragraph, _ := regexp.Compile("[ ]{2}")
	string = paragraph.ReplaceAllString(string, "\n\n")

	return string
}

func S(count int) string {
	if count == 1 {
		return ""
	}

	return "s"
}

func Join(list []string) string {
	if len(list) == 0 {
		return "None"
	}
	return strings.Join(list, ", ")
}

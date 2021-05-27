package slack_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	kac "github.com/nais/kolide-check-validator/pkg/kolide-api-client"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func New(client *http.Client, slackWebhook string) *SlackClient {
	return &SlackClient{
		slackWebhook: slackWebhook,
		client:       client,
	}
}

func (sc *SlackClient) Notify(ctx context.Context, checks []kac.Check) error {
	if len(checks) == 0 {
		return fmt.Errorf("no checks")
	}

	body, err := getRequestBody(checks)
	if err != nil {
		return fmt.Errorf("get request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sc.slackWebhook, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			bytes = []byte("unable to read error response body")
		}

		return fmt.Errorf("unable to notify Slack: HTTP %d: %s", resp.StatusCode, bytes)
	}

	log.Info("Slack has been notified")

	return nil
}

func getRequestBody(checks []kac.Check) ([]byte, error) {
	var blocks []Block

	for _, check := range checks {
		blocks = append(blocks, Block{
			Type: "divider",
		}, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: fmt.Sprintf(
					"*<https://k2.kolide.com/1401/checks/%d|%s>* - *%d failure%s*\n%s\n\nCompatibility: _%s_, Topics: _%s_",
					check.Id,
					check.Name,
					check.FailingDeviceCount,
					s(check.FailingDeviceCount),
					mrkdown(check.Description),
					na(check.Compatibility),
					na(check.Topics),
				),
			},
		})
	}

	body, err := json.Marshal(&Message{
		Blocks: append([]Block{
			{
				Type: "header",
				Text: &Text{
					Type:  "plain_text",
					Text:  ":warning: The following Kolide checks are missing severity tags: :warning:",
					Emoji: b(true),
				},
			},
		}, blocks...),
	})

	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	return body, nil
}

func b(b bool) *bool {
	return &b
}

func mrkdown(string string) string {
	bold, _ := regexp.Compile("\\*\\*")
	string = bold.ReplaceAllString(string, "*")

	links, _ := regexp.Compile("\\[(.*?)\\]\\((.*?)\\)")
	string = links.ReplaceAllString(string, "<$2|$1>")

	paragraph, _ := regexp.Compile("  ")
	string = paragraph.ReplaceAllString(string, "\n\n")

	return string
}

func s(count int) string {
	if count == 1 {
		return ""
	}

	return "s"
}

func na(list []string) string {
	joined := strings.Join(list, ", ")

	if joined == "" {
		return "N/A"
	}

	return joined
}

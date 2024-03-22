package slack_client

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type SlackClient struct {
	slackWebhook string
	client       *http.Client
	log          logrus.FieldLogger
}

type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji *bool  `json:"emoji,omitempty"`
}

type Block struct {
	Type     string `json:"type"`
	Text     *Text  `json:"text,omitempty"`
	Elements []Text `json:"elements,omitempty"`
}

type Message struct {
	Blocks []Block `json:"blocks"`
}

package kolide_api_client

import (
	"net/http"
	"strings"
)

type KolideClient struct {
	baseUrl string
	client  *http.Client
}

type Pagination struct {
	Next          string `json:"next"`
	NextCursor    string `json:"next_cursor"`
	CurrentCursor string `json:"current_cursor"`
	Count         int    `json:"count"`
}

type Check struct {
	Id                 int      `json:"id"`
	Name               string   `json:"name"`
	FailingDeviceCount int      `json:"failing_device_count"`
	Tags               []string `json:"tags"`
	Description        string   `json:"description"`
	Compatibility      []string `json:"compatibility"`
	Topics             []string `json:"topics"`
}

func (c *Check) HasSeverityTag() bool {
	if c == nil {
		return false
	}

	severityTags := []string{"info", "notice", "warning", "danger", "critical"}
	for _, tag := range c.Tags {
		tag = strings.ToLower(tag)
		for _, severityTag := range severityTags {
			if tag == severityTag {
				return true
			}
		}
	}

	return false
}

type ChecksResponse struct {
	Checks     []Check    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

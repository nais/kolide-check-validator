package kolide_api_client

import (
	"net/http"
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

type ChecksResponse struct {
	Checks     []Check    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

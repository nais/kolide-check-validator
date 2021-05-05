package kolide_api_client

import "net/http"

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
	Tags []string `json:"tags"`
}

type ChecksResponse struct {
	Checks     []Check    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

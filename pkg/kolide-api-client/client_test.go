package kolide_api_client

import (
	"net/http"
	"testing"
	"time"
)

func Test_getRetryAfter(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    time.Duration
	}{
		{
			name:    "no headers",
			headers: nil,
			want:    0,
		},
		{
			name: "no retry-after header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: 0,
		},
		{
			name: "retry-after as seconds",
			headers: http.Header{
				"Retry-After": []string{"123"},
			},
			want: 123 * time.Second,
		},
		{
			name: "retry-after as date string in the past",
			headers: http.Header{
				"Retry-After": []string{"Mon, 02 Jan 2006 15:04:05 MST"},
			},
			want: DefaultRetryAfter,
		},
		{
			name: "retry-after as unsupported string",
			headers: http.Header{
				"Retry-After": []string{"foobar"},
			},
			want: DefaultRetryAfter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRetryAfter(tt.headers); got != tt.want {
				t.Errorf("getRetryAfter(%v) = %v, want %v", tt.headers, got, tt.want)
			}
		})
	}
}

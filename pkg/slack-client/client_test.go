package slack_client

import "testing"

func Test_mrkdown(t *testing.T) {
	tests := []struct {
		name string
		input string
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
			name:   "paragraphs",
			input:  "some  new  paragraph.",
			output: `some

new

paragraph.`,
		},
		{
			name:   "all converesions",
			input:  "This **is** a [link](url).  New paragraph.",
			output: `This *is* a <url|link>.

New paragraph.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mrkdown(tt.input); got != tt.output {
				t.Errorf("mrkdown() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_s(t *testing.T) {
	tests := []struct {
		name string
		input int
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
				t.Errorf("s() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_na(t *testing.T) {
	tests := []struct {
		name string
		input []string
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
				t.Errorf("na(%v) = %v, want %v", tt.input, got, tt.output)
			}
		})
	}
}

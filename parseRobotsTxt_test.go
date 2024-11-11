package main

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRobotsTxt(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected Records
	}{
		{
			name: "dracula",
			url:  "top10/Dracula | Project Gutenberg/robots.txt",
			expected: Records{
				Records: []Record{
					{
						userAgents: []string{"*"},
						allow:      nil,             
						disallow:   []string{"*chap21.html"},
						crawlDelay: time.Duration(1) * time.Second,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := MockServerHandler()
			defer server.Close()

			actual, err := ParseRobotsTxt(server.URL + "/" + test.url)
			if err != nil {
				t.Errorf("ERROR parsing robots.txt: %v", err)
			}

			if !reflect.DeepEqual(test.expected.Records, actual.Records) {
				t.Errorf("ERROR\nExpected: %v\nActual:   %v", test.expected.Records, actual.Records)
			}
		})
	}
}

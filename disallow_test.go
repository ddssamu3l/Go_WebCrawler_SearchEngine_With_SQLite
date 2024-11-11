package main

import (
	"net/url"
	"testing"
	"time"
)

func TestDisallow(t *testing.T){
	tests := []struct{
		name string
		searchWord string
		seed string
		expected string
		expectedRecords Records
	}{
		{
			name: "Disallowing chap21.html",
			seed: "top10/Dracula | Project Gutenberg/index.html",
			searchWord: "blood",
			expected: "top10/Dracula | Project Gutenberg/chap10.html", // the expected most relevant result should not be "chap21.html" because it's disallowed by the robots.txt
			expectedRecords: Records{
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

	server := MockServerHandler()
	defer server.Close()

	for _, test := range(tests){
		t.Run(test.name, func (t *testing.T){
			idx := &InvertedIndex{
				idx: make(map[string]freq),
				docWordCountMap: make(map[string]*docResult),
			}

			Crawl(idx, server.URL + "/" + test.seed)

			actual, err := TfIdf(idx, test.searchWord)
			if err != nil{
				t.Errorf("%v\n", err)
			}
			path, err := url.PathUnescape(actual.filepath)
			if err != nil {
				t.Fatalf("ERROR: Failed to decode actual result: %v\n", err)
			}

			if path != test.expected{
				t.Errorf("ERROR: Case %s\nExpected: %s\nActual: %s\n\n", test.name, test.expected, actual.filepath)
			}
		})
	}
}
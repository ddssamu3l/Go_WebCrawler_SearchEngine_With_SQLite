package main

import (
	// "io"
	// "net/http"
	// "net/http/httptest"
	// "os"
	"reflect"
	//"strings"
	"testing"

	"github.com/kljensen/snowball"
)

func TestSearch(t *testing.T){
	tests := []struct{
		name string
		seed string
		searchWord string
		expected map[string]int
	}{
		{
			name: "Case: sceneI_30.0.html",
			seed: "rnj_files/sceneI_30.0.html",
			searchWord: "Verona",
			expected: map[string]int{
				"rnj_files/sceneI_30.0.html": 1,
			},
		},{
			name: "Case: sceneI_30.1.html",
			seed: "rnj_files/sceneI_30.1.html",
			searchWord: "Benvolio",
			expected: map[string]int{
				"rnj_files/sceneI_30.1.html": 26,
			},
		},{
			name: "Case: index.html",
			seed: "rnj_files/index.html",
			searchWord: "Romeo",
			expected: map[string]int{
				"rnj_files/index.html": 200,
				"rnj_files/sceneI_30.0.html":  2,
				"rnj_files/sceneI_30.1.html":  22,
				"rnj_files/sceneI_30.3.html":  2,
				"rnj_files/sceneI_30.4.html":  17,
				"rnj_files/sceneI_30.5.html":  15,
				"rnj_files/sceneII_30.2.html": 42,
				"rnj_files/sceneI_30.2.html":  15,
				"rnj_files/sceneII_30.0.html": 3,
				"rnj_files/sceneII_30.1.html": 10,
				"rnj_files/sceneII_30.3.html": 13,
			},
		},
	}

	server := MockServerHandler()
	defer server.Close()

	for _, test := range tests{
		t.Run(test.name, func(t *testing.T) {
			// generate expected results:
			expectedResults := make(map[string]int)
			for key, val := range test.expected {
				expectedResults[key] = val
			}

			// Initialize the inverted index
			idx := &InvertedIndex{
				idx:          make(map[string]freq),
				docWordCountMap: make(map[string]*docResult),
			}
			Crawl(idx, server.URL + "/" + test.seed)

			// adding the mock server's url to the url provided in the test case
			stemmedSearchWord, err := snowball.Stem(test.searchWord, "english", true)
			if err != nil {
				t.Errorf("ERROR stemming searchWord: %s, Stem() returned: %v", test.searchWord, err)
			}
			actual, _, err :=  idx.Lookup(stemmedSearchWord)
			if err != nil {
				t.Errorf("ERROR: Search() returned \n%v\n", err)
			}
			
			if !reflect.DeepEqual(expectedResults, actual) {
				t.Errorf("\nERROR with %s\n Expected: %v\nActual: %v\n", test.name, expectedResults, actual)
			}
		})
	}
}
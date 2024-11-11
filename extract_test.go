package main

import (
	"reflect"
	"sync"
	"testing"
)

func TestExtract(t *testing.T){
	tests := []struct{
		name string
		doc []byte
		words []string
		hrefs []string
	}{
		{
			name: "General Case",
			doc: []byte(`<!DOCTYPE html><html><body>
					<a href = "/something">sth<a>
					<a href = "/another">another<a>
					<a href = "https://www.example.com/">example<a>
				</body>
			</html>`), 
			words: []string{"sth", "another", "example"},
			hrefs: []string{"/something", "/another", "https://www.example.com/"},
		},{
			name: "Lab02 Example Case",
			doc: []byte(`<!DOCTYPE html>
			<html>
				<head>
					<title>CS272 | Welcome</title>
				</head>
				<body>
					<p>Hello World!</p>
					<p>Welcome to <a href="https://cs272-f24.github.io/">CS272</a>!</p>
				</body>
			</html>`),
			words: []string{"CS272", "Welcome", "Hello", "World", "Welcome", "to", "CS272"},
			hrefs: []string{"https://cs272-f24.github.io/"},
		},{
			name: "Blank Case",
			doc: []byte(``),
			words: nil,
			hrefs: nil,
		},
	}	

	for _, test := range tests{
		extractWg := &sync.WaitGroup{}
		extractChan := make(chan ExtractResult, 1)
		extractWg.Add(1)
		go func(doc []byte, url string, wg *sync.WaitGroup, ch chan ExtractResult){
			Extract(doc, url, wg, ch)
		}(test.doc, "http://example.com", extractWg, extractChan)

		extractWg.Wait()
		close(extractChan)

		for result := range extractChan{
			if result.err != nil{
				t.Fatalf("ERROR. Extract(%v) returned error: %v", test.doc, result.err)
			}

			if !reflect.DeepEqual(result.words, test.words){
				t.Errorf("Failed Test Case: %s\nExpected: %q\nActual: %q\n\n", test.name, test.words, result.words)
			}
			if !reflect.DeepEqual(result.hrefs, test.hrefs){
				t.Errorf("Failed Test Case: %s\nExpected: %q\nActual: %q\n\n", test.name, test.hrefs, result.hrefs)
			}
		}
	}
}


package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

func TestDownload(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected []byte
	}{
		{
			name: "index.html case",
			url:  "/index.html",
			expected: []byte(`<html>
			<body>
			   <ul>
				  <li>
					 <a href="/tests/project01/simple.html">simple.html</a>
				  </li>
				  <li>
					 <a href="/tests/project01/href.html">href.html</a>
				  </li>
				  <li>
					 <a href="/tests/project01/style.html">style.html</a>
			   </ul>
			</body>`),
		},
		{
			name: "href.html case",
			url:  "/href.html",
			expected: []byte(`<html>
			<body>
			For a simple example, see <a href="/tests/project01/simple.html">simple.html</a>
			</body>
			</html>`),
		},
		{
			name: "simple.html case",
			url:  "/simple.html",
			expected: []byte(`<html>
			<body>
			Hello CS 272, there are no links here.
			</body>
			</html>`),
		},
		{
			name: "style.html case",
			url:  "/style.html",
			expected: []byte(`<html>
			<head>
			  <title>Style</title>
			  <style>
				a.blue {
				  color: blue;
				}
				a.red {
				  color: red;
				}
			  </style>
			<body>
			  <p>
				Here is a blue link to <a class="blue" href="/href.html">href.html</a>
			  </p>
			  <p>
				And a red link to <a class="red" href="/simple.html">simple.html</a>
			  </p>
			</body>
			</html>`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == test.url {
					w.Write(test.expected)
				}
			})

			server := httptest.NewServer(handler)
			defer server.Close()

			var wg sync.WaitGroup
			ch := make(chan DownloadContent, 1)
			wg.Add(1)
			Download(server.URL+test.url, ch, &wg)
			actual := <-ch
			wg.Wait()

			if !reflect.DeepEqual(test.expected, actual.body) {
				t.Errorf("\nERROR with %s\n Expected: %s\nActual: %s\n", test.name, string(test.expected), string(actual.body))
			}
		})
	}
}

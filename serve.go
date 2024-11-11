package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type PageData struct {
    FilePath string
    Title    string
}

func Serve(idx Index) {
	// Serve static files
	http.Handle("/", http.FileServer(http.Dir("static")))

	http.Handle("/top10/", http.StripPrefix("/top10/", http.FileServer(http.Dir("./top10"))))

	// Handle search requests
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// check if the user entered a stop-word as the search input
		stopWords, err := GenerateStopWords()
		if err != nil{
			fmt.Printf("ERROR Generating stop words: %v\n", err)
		}
		searchTerm := r.URL.Query().Get("searchword")

		// if the user entered a stop word as an input, just reject the search
		if Stop(searchTerm, stopWords){
			w.Write([]byte("You have entered a Stop-Word: " + searchTerm + ". Please search again with a more relevant word. "))
			return
		}

		// Find the search result
		actual, err := TfIdf(idx, searchTerm)
		if err != nil {
			fmt.Println("ERROR with search:", err)
		}
		if !strings.HasPrefix(actual.filepath, "/") {
			actual.filepath = "/" + actual.filepath
		}

		// Display the search result
		//w.Write([]byte("Most relevant document is: " + result))
		data := PageData{
            FilePath: actual.filepath,
            Title:    actual.title,
        }

		tmpl := `<!DOCTYPE HTML>
		<html>
		<head>
			<title>Search Result</title>
		</head>
		<body>
			The most relevant page is: <a href="https://www.usfca.edu{{.FilePath}}">{{.Title}}</a>
		</body>
		</html>`
		
		// Parse the template
		t, err := template.New("searchResult").Parse(tmpl)
		if err != nil {
			fmt.Printf("ERROR: Failed to parse template: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Execute the template with data
		err = t.Execute(w, data)
		if err != nil {
			fmt.Printf("ERROR: Failed to execute template: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func MockServerHandler() *httptest.Server {
	// Mock server serving the expected response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trim leading '/' from r.URL.Path
		filePath := strings.TrimPrefix(r.URL.Path, "/")

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		w.Write(fileContent)
	})

	server := httptest.NewServer(handler)
	return server
}

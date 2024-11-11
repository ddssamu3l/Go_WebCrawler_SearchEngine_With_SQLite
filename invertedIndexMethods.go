package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kljensen/snowball"
)

func InitializeInMemoryIndex(makeNewDatabase bool) *InvertedIndex{
	if makeNewDatabase {
		fmt.Println("Error: memory mode cannot accept makeNewDatabase flag. Please re-run without -makeNewDatabase.")
		flag.Usage()
		os.Exit(1)
	}
	idx := &InvertedIndex{
		idx:             make(map[string]freq),
		docWordCountMap: make(map[string]*docResult),
	}
	return idx
}

func (idx *InvertedIndex) AddToIndex(stopwords StopWords, words []string, currentURL, docTitle string){
	// Record the number of words inside of a particular document
	doc, exists := idx.docWordCountMap[currentURL]
	if !exists{
		idx.docWordCountMap[currentURL] = &docResult{docTitle, len(words)}
	}else{
		doc.wordCount += len(words)
		doc.title = docTitle
	}

	// Add the extracted words into the inverted index
	for _, word := range words {
		word = strings.ToLower(word)

		// Check if the word is a stop word. If it is a stop word, skip to the next for loop run.
		if Stop(word, stopwords) {
			continue
		}
		// Stem the word
		stemmedWord, err := snowball.Stem(word, "english", true)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Retrieve the urlMap for the stemmed word
		urlMap, exists := idx.idx[stemmedWord]
		// If the word does not exist in the inverted index map, make a new hashmap and add it to the inverted index
		if !exists {
			urlMap = make(map[string]int)
			idx.idx[stemmedWord] = urlMap
		}
		// Increment the frequency of the current word in the current URL by 1
		urlMap[currentURL]++
	}
}

func (idx *InvertedIndex) Lookup(searchWord string)(map[string]int, map[string]*docResult, error){
	frequencyMap := idx.idx[searchWord]
	if len(frequencyMap) == 0 {
		return nil, nil, nil
	}
	docWordCount := idx.docWordCountMap
	return frequencyMap, docWordCount, nil
}
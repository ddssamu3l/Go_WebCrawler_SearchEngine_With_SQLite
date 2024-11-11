package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

/*
   Generate a hashset of stop words by reading from a JSON file
*/
func GenerateStopWords() (StopWords, error) {
	// Read the Stop words from "stopwords-en.json" and generate a stop word map
	fileContent, err := ioutil.ReadFile("stopwords-en.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	var stopwords []string

	err = json.Unmarshal(fileContent, &stopwords)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil, err
	}

	// Make the set
	set := make(StopWords)
	for _, word := range stopwords {
		set[word] = struct{}{}
	}
	return set, nil
}

/*
	Stop() will return stop words that are "meaningless", and stop words will not be added to the inverted index
	return true if a search word is a stop word
	return false if a search word is not a stop word
*/
func Stop(searchWord string, set map[string]struct{}) bool{
	_, exists := set[searchWord]
	return exists;
}
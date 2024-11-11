package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kljensen/snowball"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func TestAddToIndex(t *testing.T) {
	tests := []struct{
		name string
		documentURL string
		words []string
        numOfNonStopWords int
		searchWord string
		expectedFrequency int
	}{
		{	
			name: "The quick brown fox jumps over the lazy dog quick quick",
			documentURL: "testing.html",
			words: []string{"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog", "quick", "quick"},
            numOfNonStopWords: 6,
			searchWord: "quick",
			expectedFrequency: 3,
		},{
			name: "One Two Three",
			documentURL: "testing.html",
			words: []string{"One", "Two", "Three", "rabbit"},
            numOfNonStopWords: 1,
			searchWord: "rabbit",
			expectedFrequency: 1, // "One" is a stop word
		},{
			name: "The early bird gets the worm",
			documentURL: "testing.html",
			words: []string{"The", "early", "bird", "gets", "the", "worm"},
            numOfNonStopWords: 2,
			searchWord: "bird",
			expectedFrequency: 1,
		},
	}

    stopWords, _ := GenerateStopWords()

    for _, test := range tests{
        t.Run(test.name, func(t *testing.T) {
            idx := &Database{}
            os.Remove("index.db")
            if err := idx.InitializeDatabases(); err != nil{
                t.Errorf("ERROR Initializing Databases")
            }
            idx.AddToIndex(stopWords, test.words, test.documentURL, "TestDoc")
            // stem the searchWord
            test.searchWord, _ = snowball.Stem(test.searchWord, "english", true)

            // verify the document word count
            var wordCount int
            if err := idx.db.QueryRow(`SELECT word_count FROM documents WHERE url = ?`, test.documentURL).Scan(&wordCount); err != nil{
                t.Errorf("ERROR verifying document word count\n")
            }
            if wordCount != len(test.words){
                t.Errorf("ERROR incorrect document wordCount. Expected: %d, Actual: %d\n", len(test.words), wordCount)
            }

            // verify the number of non-stop words
            var numOfNonStopWords int
            if err := idx.db.QueryRow(`SELECT COUNT (*) from words`).Scan(&numOfNonStopWords); err != nil{
                t.Errorf("ERROR counting the number of rows in words table: %v\n", err)
            }
            if numOfNonStopWords != test.numOfNonStopWords{
                t.Errorf("ERROR with case: %v\n Expected number of non stop words: %d, Actual: %d\n", test.name, test.numOfNonStopWords, numOfNonStopWords)
            }

            // verify the frequencies table
            // find the wordID and the documentID
            var wordID, documentID int
            idx.db.QueryRow(`SELECT id FROM words WHERE word = ?`, test.searchWord).Scan(&wordID)
            idx.db.QueryRow(`SELECT id FROM documents WHERE url = ?`, test.documentURL).Scan(&documentID)
            var frequency int
            if err := idx.db.QueryRow(`SELECT frequency FROM frequencies WHERE word_id = ? AND document_id = ?`, wordID, documentID).Scan(&frequency); err != nil{
                t.Errorf("ERROR Selecting frequency from frequencies table: %v\n", err)
            }
            if frequency != test.expectedFrequency{
                t.Errorf("ERROR with case: %s\n Expected frequency for word: %s is: %d. Actual frequency: %d\n", test.name, test.searchWord, test.expectedFrequency, frequency)
            }
            idx.db.Close()
            os.Remove("index.db")
        })
    }
	
} 

func TestLookup(t *testing.T){
    tests := []struct{
        name string
        documentURL string
        words []string
        searchWord string
        expectedFrequencyMap map[string]int
        expectedDocWordCountMap map[string]*docResult
    }{
        {
            name: "The quick brown fox jumps over the lazy dog quick quick",
			documentURL: "testing.html",
			words: []string{"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog", "quick", "quick"},
            searchWord: "fox",
            expectedFrequencyMap: map[string]int{
                "testing.html": 1,
            },
            expectedDocWordCountMap: map[string]*docResult{
                "testing.html": &docResult{
                    title:     "Test",
                    wordCount: 11,
                },
            },            
        },{
            name: "One Two Three",
			documentURL: "testing2.html",
			words: []string{"One", "Two", "Three", "rabbit"},
            searchWord: "rabbit",
            expectedFrequencyMap: map[string]int{
                "testing2.html": 1,
            },
            expectedDocWordCountMap: map[string]*docResult{
                "testing2.html": &docResult{
                    title:     "Test",
                    wordCount: 4,
                },
            },
        },{
            name: "The early bird gets the worm",
			documentURL: "testing3.html",
			words: []string{"The", "early", "bird", "gets", "the", "worm"},
			searchWord: "bird",
            expectedFrequencyMap: map[string]int{
                "testing3.html": 1,
            },
            expectedDocWordCountMap: map[string]*docResult{
                "testing3.html": &docResult{
                    title:     "Test",
                    wordCount: 6,
                },
            },
        },
    }

    stopWords, _ := GenerateStopWords()

    for _, test := range tests{
        t.Run(test.name, func(t *testing.T){
            idx := &Database{}
            os.Remove("index.db")
            if err := idx.InitializeDatabases(); err != nil{
                t.Errorf("ERROR Initializing Databases")
            }
            idx.AddToIndex(stopWords, test.words, test.documentURL, "Test")
            // stem the searchWord
            test.searchWord, _ = snowball.Stem(test.searchWord, "english", true)

            frequencyMap, docWordCountMap, err := idx.Lookup(test.searchWord)
            if err != nil{
                t.Errorf("ERROR with Lookup(): %v\n", err)
            }

            if !reflect.DeepEqual(test.expectedFrequencyMap, frequencyMap){
                t.Errorf("ERROR with case: %s\n Expected frequencyMap: %v\n Actual FrequencyMap: %v\n", test.name, test.expectedFrequencyMap, frequencyMap)
            }
            if !reflect.DeepEqual(test.expectedDocWordCountMap, docWordCountMap){
                t.Errorf("ERROR with case: %s\n Expected docWordCountMap: %v\n Actual docWordCountMap: %v\n", test.name, test.expectedDocWordCountMap[test.documentURL], docWordCountMap[test.documentURL])
            }

            idx.db.Close()
            os.Remove("index.db")
        })
    }
}

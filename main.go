package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"
    "time"
)

type Index interface {
    AddToIndex(stopwords StopWords, words []string, currentURL, docTitle string)
    Lookup(word string) (map[string]int, map[string]*docResult, error) // returns the frequency and docWordCount maps
}

type freq map[string]int

// InvertedIndex holds the index and document word counts
type InvertedIndex struct {
    idx             map[string]freq
    docWordCountMap map[string]*docResult
}

type Database struct {
    db *sql.DB
}

func main() {
    mode, makeNewDatabase := parseFlags()
    initialDatabaseExists := FileExists("index.db")

    idx, db := initializeIndex(mode, makeNewDatabase, initialDatabaseExists)

	// try to close the database after the the program ends if we made one. Error checks to prevent mistakes
    if db != nil {
        defer func() {
            if db.db != nil {
                if err := db.db.Close(); err != nil {
                    log.Printf("Error closing database: %v\n", err)
                }
            }
        }()
    }

    // GO routines for the search engine.
    go func() {
        // If the mode is "memory", we crawl
        // If the database (index.db) does not exist, we crawl
        // if the user entered -makeNewDatabase, we crawl
        if mode == "memory" || !initialDatabaseExists || makeNewDatabase {
            t1 := time.Now()
            // https://cs272-f24.github.io/top10/
            // http://localhost:8080/top10/index.html
            if err := Crawl(idx, "https://www.nyu.edu/"); err != nil {
                fmt.Println("Crawl error:", err)
            }
            t2 := time.Now()
            fmt.Println("Successfully crawled the corpus in: ", t2.Sub(t1))
        }
    }()
    go Serve(idx)

    for {
        time.Sleep(100)
    }
}

// fileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// parseFlags parses command-line flags and returns the mode and makeNewDatabase values
func parseFlags() (string, bool) {
    mode := flag.String("mode", "", "Mode of operation: In-Memory-Index/Database (required)")
    makeNewDatabase := flag.Bool("makeNewDatabase", false, "Choose to re-crawl the corpus to make a new database (optional)")
    flag.Parse()

    // Check if the user has entered a mode. If not, exit the program
    if *mode == "" {
        fmt.Println("Error: -mode flag is required.")
        flag.Usage()
        os.Exit(1)
    }

    return *mode, *makeNewDatabase
}

// initializeIndex initializes the idx variable based on the mode and returns it along with the database (if applicable)
func initializeIndex(mode string, makeNewDatabase bool, initialDatabaseExists bool) (Index, *Database) {
    var idx Index
    var db *Database

    if mode == "memory" {
        idx = InitializeInMemoryIndex(makeNewDatabase)
    } else if mode == "database" {
		// If making a new database, delete the old index.db if it exists
		if makeNewDatabase && initialDatabaseExists {
			if err := os.Remove("index.db"); err != nil {
				fmt.Println("ERROR deleting index.db:", err)
				os.Exit(1)
			}
		}
		// Initialize Database
		db = &Database{}
		idx = db

		// Initialize the databases and tables
		if err := db.InitializeDatabases(); err != nil {
			log.Fatalf("Failed to initialize databases: %v", err)
		}
    } else {
        fmt.Println("Error: unexpected input for -mode. Expected inputs for -mode:\n  memory\n  database")
        flag.Usage()
        os.Exit(1)
    }

    return idx, db
}

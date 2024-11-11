package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/kljensen/snowball"
)

type docResult struct{
    title string
    wordCount int
}

/*
	AddToIndex adds the crawl results to the SQLite database, updating the 'words', 'documents' and 'frequencies' data tables.
*/
func (idx *Database) AddToIndex(stopwords StopWords, words []string, currentURL, docTitle string) {
    tx, err := idx.db.Begin()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return
    }
    defer tx.Commit()

    // Insert the new document into the 'documents' table
    insertIntoDocuments, err := tx.Prepare(`INSERT INTO documents (url, title, word_count) VALUES (?, ?, ?)`)
    if err != nil{
        log.Printf("ERROR preparing database insertion statement: %v\n", err)
        return
    }
    if _, err := insertIntoDocuments.Exec(currentURL, docTitle, len(words)); err != nil{
        log.Printf("ERROR inserting into databases table: %v\n%s, %s, %d", err, currentURL, docTitle, len(words))
        return
    }
    var documentID int
    if err = tx.QueryRow("SELECT id FROM documents WHERE url = ?", currentURL).Scan(&documentID); err != nil {
        log.Printf("Failed to insert document: %v", err)
        return
    }

    // Prepare statements for batch insertions
    insertWordStatement, err := tx.Prepare("INSERT INTO words(word) VALUES (?) ON CONFLICT(word) DO UPDATE SET word=word")
    if err != nil {
        log.Printf("Failed to prepare insertWordStatement: %v", err)
        return
    }
    defer insertWordStatement.Close()

    insertFrequencyStatement, err := tx.Prepare(`
        INSERT INTO frequencies (word_id, document_id, frequency)
        VALUES (?, ?, ?)
    `)
    if err != nil {
        log.Printf("Failed to prepare insertFrequencyStatemet: %v", err)
        return
    }
    defer insertFrequencyStatement.Close()

    // Process words and track frequencies
    wordFrequencyMap := make(map[string]int)
    for _, word := range words {
        word = strings.ToLower(word)
        if Stop(word, stopwords) {
            continue
        }

        var stemmedWord string
        stemmedWord, err = snowball.Stem(word, "english", true)
        if err != nil {
            log.Printf("Stemming error for word '%s': %v", word, err)
            continue
        }

        wordFrequencyMap[stemmedWord]++
    }

    // insert words and frequencies
    for word, freq := range wordFrequencyMap {
        // insert the word into the words table and get that word's ID
        _, err = insertWordStatement.Exec(word)
        if err != nil {
            log.Printf("ERROR inserting word into words table: %v\n", err)
        }
        var wordID int
        if err = tx.QueryRow(`SELECT id FROM words WHERE word = ?`, word).Scan(&wordID); err != nil {
            log.Printf("Error inserting word '%s': %v", word, err)
            return
        }

        // insert or update the frequency table
        if _, err = insertFrequencyStatement.Exec(wordID, documentID, freq); err != nil {
            log.Printf("Failed to insert frequency for word '%s': %v", word, err)
            return
        }
    }
}

func (idx *Database) Lookup(searchWord string) (map[string]int, map[string]*docResult, error) {
    // Start the db transaction
    tx, err := idx.db.Begin()
    if err != nil {
        return nil, nil, fmt.Errorf("Failed to begin transaction: %v", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()

    // Convert search word to lowercase and stem it
    searchWord = strings.ToLower(searchWord)
    stemmedWord, err := snowball.Stem(searchWord, "english", true)
    if err != nil {
        return nil, nil, fmt.Errorf("Error stemming search word '%s': %v", searchWord, err)
    }

    // Retrieve the 'word_id' of the stemmedWord from the 'words' table
    var wordID int
    if err = tx.QueryRow("SELECT id FROM words WHERE word = ?", stemmedWord).Scan(&wordID); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil, nil
        }
        return nil, nil, fmt.Errorf("Failed to retrieve word ID for '%s': %v", searchWord, err)
    }

    // retrieve frequencies and word counts in a single query using the transaction to reduce repeated queries, speeding up performance
    rows, err := tx.Query(`
        SELECT documents.url, frequencies.frequency, documents.word_count
        FROM frequencies
        JOIN documents ON frequencies.document_id = documents.id
        WHERE frequencies.word_id = ?
    `, wordID)
    if err != nil {
        return nil, nil, fmt.Errorf("Error querying frequencies: %v", err)
    }
    defer rows.Close()

	// build frequencyMap to track how many times the searchWord appears in each document that contains the searchWord
    frequencyMap := make(map[string]int)
    for rows.Next() {
        var url string
        var frequency, wordCount int
        if err = rows.Scan(&url, &frequency, &wordCount); err != nil {
            return nil, nil, fmt.Errorf("Error scanning row: %v", err)
        }
        frequencyMap[url] = frequency
    }
    if err = rows.Err(); err != nil {
        return nil, nil, fmt.Errorf("Row iteration error: %v", err)
    }

	// build docWordCountMap by finding the word count of EVERY document in the database
    docWordCountMap := make(map[string]*docResult)
    rows, err = tx.Query(`
        SELECT url, title, word_count
        FROM documents
    `)
    if err != nil {
        return nil, nil, fmt.Errorf("Error querying documents: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var url string
        var title string
        var wordCount int
        if err = rows.Scan(&url, &title, &wordCount); err != nil {
            return nil, nil, fmt.Errorf("Error scanning document row: %v", err)
        }
        docWordCountMap[url] = &docResult{title, wordCount}
    }
    if err = rows.Err(); err != nil {
        return nil, nil, fmt.Errorf("Row iteration error: %v", err)
    }

    return frequencyMap, docWordCountMap, nil
}


func (idx *Database) InitializeDatabases() error{
    // Initialize a new database connection
    db, err := sql.Open("sqlite3", "index.db")
    if err != nil {
        return fmt.Errorf("%v\n", err)
    }
    idx.db = db

    // Enable foreign key constraints
    _, err = db.Exec("PRAGMA foreign_keys = ON;")
    if err != nil {
        db.Close()
        return fmt.Errorf("%v\n", err)
    }

    // Initialize the 'words' table
    sqlCreateWordsTable := `
        CREATE TABLE IF NOT EXISTS words (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            word TEXT UNIQUE
        );
    `
    stmt, err := db.Prepare(sqlCreateWordsTable)
    if err != nil {
        log.Printf("Prepare returned error for 'words' table: %v", err)
        return fmt.Errorf("%v\n", err)
    }
    stmt.Exec()
	stmt.Close()
    

    // Initialize the 'documents' table
    sqlCreateDocumentsTable := `
        CREATE TABLE IF NOT EXISTS documents (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT UNIQUE,
            title TEXT,
            word_count INTEGER
        );
    `
    stmt, err = db.Prepare(sqlCreateDocumentsTable)
    if err != nil {
        log.Printf("Prepare returned error for 'documents' table: %v", err)
        return fmt.Errorf("%v\n", err)
    }
	stmt.Exec()
	stmt.Close()

    // Initialize the 'frequencies' table
    sqlCreateFrequenciesTable := `
        CREATE TABLE IF NOT EXISTS frequencies (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            word_id INTEGER,
            document_id INTEGER,
            frequency INTEGER,
            FOREIGN KEY (word_id) REFERENCES words(id) ON DELETE CASCADE,
            FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
			UNIQUE(word_id, document_id)
        );
    `
    stmt, err = db.Prepare(sqlCreateFrequenciesTable)
    if err != nil {
        log.Printf("Prepare returned error for 'frequencies' table: %v", err)
        return fmt.Errorf("%v\n", err)
    }
	stmt.Exec()
	stmt.Close()

    return nil
}

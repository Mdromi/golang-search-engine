package indexer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// InMemoryBoltDB is an in-memory implementation of *bolt.DB for testing.
type InMemoryBoltDB struct {
	*bolt.DB
	data map[string][]byte
}

// FindProjectRoot traverses the file system upwards to find the project root directory.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Replace "data" with a directory or file name that indicates the project root.
	for {
		if _, err := os.Stat(filepath.Join(dir, "data")); err == nil {
			return dir, nil
		}

		// Stop the traversal if we have reached the root directory ("/" on Unix systems).
		if dir == filepath.Dir(dir) {
			break
		}

		// Move one level up in the file system hierarchy.
		dir = filepath.Dir(dir)
	}

	// Return an error if we couldn't find the project root.
	return "", fmt.Errorf("project root directory not found")
}

// NewInMemoryBoltDB creates a new instance of InMemoryBoltDB.
func NewInMemoryBoltDB() (*InMemoryBoltDB, func() error) {
	// Get the project root directory
	projectRoot, err := FindProjectRoot()
	if err != nil {
		fmt.Println("Failed to find project root:", err)
		panic(err)
	}

	// Construct the file path relative to the project root directory
	filePath := filepath.Join(projectRoot, "data", "mydb.db")

	// Create an in-memory BoltDB instance using bolt.Open.
	// Note: You may want to use other bolt.Open options for testing purposes.
	db, err := bolt.Open(filePath, 0666, &bolt.Options{})

	if err != nil {
		fmt.Println("Failed to create in-memory BoltDB:", err)
		panic("Failed to create in-memory BoltDB")
	}

	inMemoryDB := &InMemoryBoltDB{
		DB:   db,
		data: make(map[string][]byte),
	}

	// Define the cleanup function to close the database.
	cleanup := func() error {
		return inMemoryDB.DB.Close()
	}

	return inMemoryDB, cleanup
}

// Update implements the Update method of the IndexDB interface.
func (db *InMemoryBoltDB) Update(fn func(*bolt.Tx) error) error {
	return fn(nil) // For an in-memory database, we can pass nil for the transaction.
}

// View implements the View method of the IndexDB interface.
func (db *InMemoryBoltDB) View(fn func(*bolt.Tx) error) error {
	return fn(nil) // For an in-memory database, we can pass nil for the transaction.
}

func (db *InMemoryBoltDB) Close() error {
	return db.DB.Close()
}

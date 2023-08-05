// indexer/database.go

package indexer

import (
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// BoltDB represents a BoltDB instance.
type BoltDB struct {
	DB *bolt.DB
}

// Database is an interface that represents the database functionality.
type Database interface {
	Update(fn func(*bolt.Tx) error) error
	View(fn func(*bolt.Tx) error) error
	Close() error
}

// NewBoltDB creates a new instance of BoltDB and opens the specified database file.
func NewBoltDB(path string) (*BoltDB, func() error, error) {
	// Get the project root directory
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Construct the file path relative to the project root directory
	filePath := filepath.Join(projectRoot, path)

	// Create a BoltDB instance using bolt.Open.
	db, err := bolt.Open(filePath, 0666, &bolt.Options{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create BoltDB: %w", err)
	}

	boltDB := &BoltDB{
		DB: db,
	}

	// Define the cleanup function to close the database.
	cleanup := func() error {
		return boltDB.DB.Close()
	}

	return boltDB, cleanup, nil
}

func (db *BoltDB) Update(fn func(*bolt.Tx) error) error {
	return db.DB.Update(fn)
}

func (db *BoltDB) View(fn func(*bolt.Tx) error) error {
	return db.DB.View(fn)
}

func (db *BoltDB) Close() error {
	return db.DB.Close()
}

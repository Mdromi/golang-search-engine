package indexer

import (
	"github.com/boltdb/bolt"
)

// InMemoryBoltDB is an in-memory implementation of *bolt.DB for testing.
type InMemoryBoltDB struct {
	*bolt.DB
	data map[string][]byte
}

// NewInMemoryBoltDB creates a new instance of InMemoryBoltDB.
func NewInMemoryBoltDB() (*InMemoryBoltDB, func() error) {
	// Create an in-memory BoltDB instance using bolt.Open.
	// Note: You may want to use other bolt.Open options for testing purposes.
	// db, err := bolt.Open(":memory:", 0666, &bolt.Options{})
	db, err := bolt.Open("", 0666, &bolt.Options{})
	if err != nil {
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

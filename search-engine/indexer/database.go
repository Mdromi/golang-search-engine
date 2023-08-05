// indexer/database.go

package indexer

import "github.com/boltdb/bolt"

// Database is an interface that represents the database functionality.
type Database interface {
	Update(fn func(tx *bolt.Tx) error) error
	View(fn func(tx *bolt.Tx) error) error
	Close() error
}

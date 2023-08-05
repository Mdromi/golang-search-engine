package indexer

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type Indexer struct {
	db    *InMemoryBoltDB
	redis *redis.Client
}

// IndexDB is an interface that represents a database for indexing.
type IndexDB interface {
	Update(func(*bolt.Tx) error) error
	View(func(*bolt.Tx) error) error
}

// NewIndexer creates a new instance of the Indexer.
func NewIndexer(db *InMemoryBoltDB, redis *redis.Client) *Indexer {
	return &Indexer{
		db:    db,
		redis: redis,
	}
}

// Index performs the indexing of words and their associated URLs in batches.
func (i *Indexer) Index(data map[string][]string) error {
	// Open a writable transaction
	return i.db.Update(func(tx *bolt.Tx) error {
		// Create or access the "IndexBucket"
		bucket, err := tx.CreateBucketIfNotExists([]byte("IndexBucket"))
		if err != nil {
			return err
		}

		// Use concurrency to index words in parallel
		var wg sync.WaitGroup
		for word, urls := range data {
			wg.Add(1)
			go func(word string, urls []string) {
				defer wg.Done()

				// Use the word as the key and the URLs as the value
				if err := bucket.Put([]byte(word), gzipCompress(urls)); err != nil {
					fmt.Println("Failed to index word:", word)
				}
			}(word, urls)
		}

		wg.Wait()
		return nil
	})
}

// Query searches for a given word and returns the associated URLs.
func (i *Indexer) Query(word string) ([]string, error) {
	// Try to get the data from Redis first
	urls, err := i.getFromRedis(word)
	if err == nil {
		return urls, nil
	}

	// If not found in Redis, try to get it from BoltDB
	urls, err = i.getFromBoltDB(word)
	if err != nil {
		return nil, err
	}

	// Save the data to Redis for future queries
	if err := i.saveToRedis(word, urls); err != nil {
		fmt.Println("Failed to save data to Redis:", err)
	}

	return urls, nil
}

// getFromBoltDB retrieves the compressed data from BoltDB.
func (i *Indexer) getFromBoltDB(word string) ([]string, error) {
	var urls []string

	// Open a read-only transaction
	err := i.db.View(func(tx *bolt.Tx) error {
		// Access the "IndexBucket"
		bucket := tx.Bucket([]byte("IndexBucket"))
		if bucket == nil {
			return nil // Bucket not found, return empty result
		}

		// Retrieve the compressed data
		data := bucket.Get([]byte(word))
		if data == nil {
			return nil // Word not found, return empty result
		}

		// Decompress the data and return the URLs
		urls = gzipDecompress(data)
		return nil
	})

	return urls, err
}

// getFromRedis retrieves the data from Redis.
func (i *Indexer) getFromRedis(word string) ([]string, error) {
	// Get the data from Redis
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	data, err := i.redis.Get(ctx, word).Result()
	if err != nil {
		return nil, err
	}

	// Decode the data and return the URLs
	var urls []string
	err = gob.NewDecoder(bytes.NewBufferString(data)).Decode(&urls)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

// saveToRedis saves the data to Redis.
func (i *Indexer) saveToRedis(word string, urls []string) error {
	// Encode the data
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(urls)
	if err != nil {
		return err
	}

	// Set the data in Redis with an expiration of 1 hour
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = i.redis.SetEX(ctx, word, buf.String(), 1*time.Hour).Result()
	return err
}

// gzipCompress compresses a slice of strings using gzip.
func gzipCompress(input []string) []byte {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(gzw)
	err := enc.Encode(input)
	gzw.Close()

	if err != nil {
		return nil
	}

	return buf.Bytes()
}

// gzipDecompress decompresses a byte slice using gzip.
func gzipDecompress(input []byte) []string {
	buf := bytes.NewBuffer(input)
	gzr, err := gzip.NewReader(buf)
	if err != nil {
		return nil
	}
	defer gzr.Close()

	var output []string
	dec := gob.NewDecoder(gzr)
	err = dec.Decode(&output)

	if err != nil {
		return nil
	}

	return output
}

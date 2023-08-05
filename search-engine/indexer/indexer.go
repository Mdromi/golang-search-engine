// indexer.go

package indexer

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Indexer struct {
	db    Database
	redis *redis.Client
}

// IndexDB is an interface that represents a database for indexing.
type IndexDB interface {
	Update(func(*bolt.Tx) error) error
	View(func(*bolt.Tx) error) error
	Close() error
}

// NewIndexer creates a new instance of the Indexer.
func NewIndexer(db Database, redis *redis.Client) *Indexer {
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
				if err := bucket.Put([]byte(word), []byte(strings.Join(urls, ","))); err != nil {
					fmt.Println("Failed to index word:", word)
				}
			}(word, append([]string{}, urls...)) // Pass the copy of "urls" to the goroutine

			// Save data to Redis after indexing in BoltDB
			if i.redis != nil {
				if err := i.saveToRedis(word, urls); err != nil {
					fmt.Println("Failed to save data to Redis:", err)
				}
			}
		}

		wg.Wait()
		return nil
	})
}

// Query searches for a given word and returns the associated URLs.
func (i *Indexer) Query(word string) ([]string, error) {
	// Try to get the data from Redis first
	if i.redis != nil {
		urls, err := i.getFromRedis(word)
		if err == nil {
			return urls, nil
		}
	}

	// If not found in Redis or Redis is not available, try to get it from BoltDB
	urls, err := i.getFromBoltDB(word)
	if err != nil {
		return nil, err
	}

	// Save the data to Redis for future queries
	if i.redis != nil {
		if err := i.saveToRedis(word, urls); err != nil {
			fmt.Println("Failed to save data to Redis:", err)
		}
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
	ctx := context.Background()
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
	// If the Redis client is nil, just return without saving
	if i.redis == nil {
		return nil
	}

	// Encode the data
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(urls)
	if err != nil {
		return err
	}

	// Set the data in Redis with an expiration of 1 hour
	ctx := context.Background()
	_, err = i.redis.Set(ctx, word, buf.String(), 1*time.Hour).Result()
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

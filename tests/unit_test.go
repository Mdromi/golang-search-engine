// tests/unit_test.go

package main_test

import (
	"testing"

	"github.com/Mdromi/golang-search-engine/search-engine/crawler"
	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	URL1 = "https://www.webscraper.io/test-sites/e-commerce/allinone"
	URL2 = "https://www.webscraper.io/test-sites/e-commerce/scroll/product/514"
	URL3 = "https://www.webscraper.io/test-sites/e-commerce/allinone-popup-links/phones"
)

// TestCrawlerCrawl tests the Crawl function of the crawler package.
func TestCrawlerCrawl(t *testing.T) {
	// Create a new crawler
	c := crawler.NewCrawler(1, 1)
	c.SetFilterDomain("voskan.host")

	// Crawl from a test URL with depth 0
	err := c.Crawl(URL1, 0)
	assert.NoError(t, err)

	// Wait for crawling to finish
	c.Wait()

	// Assert that the collected data is not empty
	collectedData := c.GetCollectedData()
	assert.NotEmpty(t, collectedData)
}

// TestIndexerIndex tests the Index function of the indexer package.
func TestIndexerIndex(t *testing.T) {
	// Create a new indexer with an in-memory BoltDB instance (for testing purposes)
	db, cleanup := indexer.NewInMemoryBoltDB()
	defer func() {
		if err := cleanup(); err != nil {
			logrus.Fatal("Failed to close in-memory BoltDB:", err)
		}
	}()

	t.Log("Creating Redis client...")
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Replace with your Redis server address
	})

	t.Log("Creating indexer...")
	idx := indexer.NewIndexer(db.DB, redisClient)

	t.Log("Preparing test data...")

	// Prepare test data
	data := map[string][]string{
		"test": []string{URL2, URL3},
	}

	// Index the test data
	err := idx.Index(data)
	if err != nil {
		t.Fatalf("Failed to index test data: %v", err)
	}

	// Verify that the data is indexed correctly
	urls, err := idx.Query("test")
	if err != nil {
		t.Fatalf("Failed to query indexed data: %v", err)
	}

	// t.Logf("Indexed URLs: %v", urls) // Log the indexed URLs

	assert.Equal(t, []string{URL2, URL3}, urls)
}

// ... Add more unit tests as needed for other packages/functions.

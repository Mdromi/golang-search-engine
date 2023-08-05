// tests/unit_test.go

package main_test

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/Mdromi/golang-search-engine/search-engine/crawler"
	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/stretchr/testify/assert"
)

const (
	URL1 = "https://voskan.host/category/golang/"
	URL2 = "https://medium.com/@mdromi/go-vs-ruby-on-rails-a-business-perspective-38a98faf2422"
	URL3 = "https://medium.com/@mdromi/discovering-golang-part-1-an-introduction-to-the-power-of-golang-662f82365d07"
)

// TestCrawlerCrawl tests the Crawl function of the crawler package.
func TestCrawlerCrawl(t *testing.T) {
	// Create a new crawler
	c := crawler.NewCrawler(1, 1)
	c.SetFilterDomain(URL1)

	// Crawl from a test URL with depth 0
	c.Crawl(URL1, 0)

	// Wait for crawling to finish
	c.Wait()

	// Assert that the collected data is not empty
	collectedData := c.GetCollectedData()
	assert.NotEmpty(t, collectedData)
}

// TestIndexerIndex tests the Index function of the indexer package.
func TestIndexerIndex(t *testing.T) {
	// Create a new indexer with an in-memory BoltDB instance (for testing purposes)
	// Set up BoltDB
	db, cleanup := indexer.NewInMemoryBoltDB()
	defer func() {
		if err := cleanup(); err != nil {
			logrus.Fatal("Failed to close in-memory BoltDB:", err)
		}
	}()
	idx := indexer.NewIndexer(db, nil)

	// Prepare test data
	data := map[string][]string{
		"test": []string{URL2, URL3},
	}

	// Index the test data
	err := idx.Index(data)
	assert.NoError(t, err)

	// Verify that the data is indexed correctly
	urls, err := idx.Query("test")
	assert.NoError(t, err)
	assert.Equal(t, []string{URL2, URL3}, urls)
}

// ... Add more unit tests as needed for other packages/functions.

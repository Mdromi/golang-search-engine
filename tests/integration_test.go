package main_test

import (
	"testing"

	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/Mdromi/golang-search-engine/search-engine/search"
	"github.com/stretchr/testify/assert"
)

func TestSearchEngineIntegration(t *testing.T) {
	t.Run("IntegrationTest", func(t *testing.T) {
		// Create a new indexer with an in-memory BoltDB instance (for testing purposes)
		db, cleanup := indexer.NewInMemoryBoltDB()
		defer cleanup()

		// idx := indexer.NewIndexer(db, nil)
		idx := indexer.NewIndexer(db.DB, nil)
		// Prepare test data
		data := map[string][]string{
			"test": []string{URL2, URL3}, // Replace URL2 and URL3 with actual URLs
		}

		// Index the test data
		err := idx.Index(data)
		assert.NoError(t, err)

		// Create a new searcher with the same in-memory BoltDB instance
		// Note: You should pass db.DB to search.NewSearcher() as it expects *bolt.DB.
		s := search.NewSearcher(db.DB)

		// Search for "test" keyword
		results, err := s.Search("test", &search.SearchOptions{})
		assert.NoError(t, err)
		assert.Equal(t, []string{URL2, URL3}, results) // Replace URL2 and URL3 with actual URLs

		// Ensure the database is closed when the test completes
		defer cleanup()
	})
}

// ... Add more integration tests as needed for other components.

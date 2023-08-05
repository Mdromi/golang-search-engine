package main_test

import (
	"fmt"
	"testing"

	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/Mdromi/golang-search-engine/search-engine/search"
	"github.com/stretchr/testify/assert"
)

func TestSearchEngineIntegration(t *testing.T) {
	t.Run("IntegrationTest", func(t *testing.T) {
		fmt.Println("STEP - 1")
		// Create a new indexer with an in-memory BoltDB instance (for testing purposes)
		db, cleanup := indexer.NewInMemoryBoltDB()
		fmt.Println("STEP - 2")
		defer cleanup() // Ensure the database is closed when the test function exits

		idx := indexer.NewIndexer(db, nil)

		// Prepare test data
		data := map[string][]string{
			"test": []string{"URL2", "URL3"}, // Replace URL2 and URL3 with actual URLs
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
		assert.Equal(t, []string{"URL2", "URL3"}, results) // Replace URL2 and URL3 with actual URLs
	})
}

// ... Add more integration tests as needed for other components.

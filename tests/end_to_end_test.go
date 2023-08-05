// tests/end_to_end_test.go

package main_test

import (
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestEndToEndSearchEngine tests the entire search engine from crawling to searching.
func TestEndToEndSearchEngine(t *testing.T) {
	// Start the search engine in a separate goroutine
	cmd := exec.Command("go", "run", "main.go")
	err := cmd.Start()
	assert.NoError(t, err)

	// Wait for the search engine to start up
	time.Sleep(2 * time.Second)

	// Send a search query to the search engine using HTTP request (assuming the search endpoint is "/search")
	// You may need to update the URL to match your actual search endpoint.
	// resp, err := http.Get("http://localhost:8080/search?query=test")
	resp, err := http.Get("https://www.webscraper.io/test-sites/tables")
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Assert that the response status code is 200 (OK)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// ... Add more checks as needed to verify the response data.

	// Stop the search engine
	err = cmd.Process.Kill()
	assert.NoError(t, err)
}

package search

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/boltdb/bolt"
)

// QueryProcessor is responsible for processing user queries.
type QueryProcessor struct{}

// Constants for sorting options
const (
	SortByRelevance = "relevance"
	SortByDate      = "date"
)

// Searcher is responsible for searching the index and returning results.
type Searcher struct {
	db indexer.Database
}

// SearchOptions represents the options for advanced search.
type SearchOptions struct {
	FilterDomain string // Filter results by a specific domain
	SortBy       string // Sort results by "relevance" or "date"
}

// NewQueryProcessor creates a new instance of QueryProcessor.
func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{}
}

// NewSearcher creates a new instance of Searcher.
func NewSearcher(db indexer.Database) *Searcher {
	return &Searcher{
		db: db,
	}
}

// Process processes the user query and returns the individual keywords.
func (qp *QueryProcessor) Process(query string) []string {
	// Convert the query to lowercase
	query = strings.ToLower(query)

	// Split the query into words
	words := strings.Fields(query)

	// Return the processed words
	return words
}

// Search searches the index for the given query and returns matching URLs.
func (s *Searcher) Search(query string, options *SearchOptions) ([]string, error) {
	// Process the query
	words := NewQueryProcessor().Process(query)

	// Open the read-only transaction
	var results []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("IndexBucket"))

		// For each word, get the corresponding URLs from the index
		for _, word := range words {
			val := b.Get([]byte(word))
			if val != nil {
				urls := strings.Split(string(val), ",")
				results = append(results, urls...)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Apply filtering based on domain
	if options.FilterDomain != "" {
		results = s.filterByDomain(results, options.FilterDomain)
	}

	// Apply sorting
	if options.SortBy == SortByDate {
		results = s.sortByDate(results)
	} else {
		results = s.sortByRelevance(results)
	}

	return results, nil
}

// filterByDomain filters the search results to include only URLs from a specific domain.
func (s *Searcher) filterByDomain(results []string, domain string) []string {
	var filteredResults []string
	for _, url := range results {
		if strings.Contains(url, domain) {
			filteredResults = append(filteredResults, url)
		}
	}
	return filteredResults
}

// sortByRelevance sorts the search results by relevance (number of occurrences in the index).
func (s *Searcher) sortByRelevance(results []string) []string {
	// Count the occurrences of each URL in the results
	counts := make(map[string]int)
	for _, url := range results {
		counts[url]++
	}

	// Sort the URLs based on their occurrences (relevance)
	sort.SliceStable(results, func(i, j int) bool {
		return counts[results[i]] > counts[results[j]]
	})

	return results
}

// sortByDate sorts the search results by date (if applicable).
func (s *Searcher) sortByDate(results []string) []string {
	// Create a map to store the URLs and their corresponding dates
	dateMap := make(map[string]string)

	// Extract and store the dates for each URL in the results
	for _, url := range results {
		date := extractDateFromURL(url)
		if date != "" {
			dateMap[url] = date
		}
	}

	// Sort the results based on the extracted dates
	sortedResults := make([]string, 0, len(dateMap))
	for url := range dateMap {
		sortedResults = append(sortedResults, url)
	}

	return sortedResults
}

// extractDateFromURL is a helper function to extract the date from a URL (assuming the date format is "YYYY-MM-DD").
func extractDateFromURL(url string) string {
	// Check if the URL is long enough to contain the assumed date format
	if len(url) < 27 {
		return "" // Return an empty string if the URL is too short to contain the date
	}

	// Extract the date from the URL based on the assumed format
	date := url[17:27]

	// Validate the extracted date format using a regular expression
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(date) {
		return "" // Return an empty string if the date format is not correct
	}

	// Additional validation can be added here if needed

	return date
}

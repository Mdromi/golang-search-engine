package crawler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Crawler is a web crawler that fetches and collects data from web pages.
type Crawler struct {
	client        *http.Client
	maxDepth      int
	concurrency   int
	rateLimiter   <-chan time.Time
	filterDomain  string
	wg            sync.WaitGroup
	collectedData *CollectedData
}

// CollectedData is a struct to represent the collected data.
type CollectedData struct {
	mutex sync.Mutex
	data  map[string][]string
}

// NewCollectedData creates a new instance of CollectedData.
func NewCollectedData() *CollectedData {
	return &CollectedData{
		data: make(map[string][]string),
	}
}

// AddData adds the URL data to the collected data.
func (cd *CollectedData) AddData(url, data string) {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	cd.data[url] = append(cd.data[url], data)
}

// GetData returns the collected data.
func (cd *CollectedData) GetData() map[string][]string {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()

	result := make(map[string][]string)
	for url, data := range cd.data {
		result[url] = data
	}
	return result
}

// NewCrawler creates a new instance of the Crawler with collected data.
func NewCrawler(maxDepth, concurrency int) *Crawler {
	return &Crawler{
		client:        &http.Client{Timeout: 30 * time.Second},
		maxDepth:      maxDepth,
		concurrency:   concurrency,
		rateLimiter:   time.Tick(500 * time.Millisecond),
		collectedData: NewCollectedData(),
	}
}

// Crawl starts the crawling process from the provided URL with the specified depth.
func (c *Crawler) Crawl(url string, depth int) error {
	// Check if we've reached the maximum depth
	if depth > c.maxDepth {
		return nil
	}

	// Throttle requests
	<-c.rateLimiter

	// Fetch the URL
	res, err := c.fetch(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Parse the page with goquery
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	// Extract the page text
	pageText := doc.Find("body").Text()

	// Process the page data (store or index the content)
	c.collectedData.AddData(url, pageText)

	// Find all links in the page and recursively crawl them
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			// Filter URLs if necessary
			if c.filterDomain == "" || strings.Contains(link, c.filterDomain) {
				c.wg.Add(1)
				go c.Crawl(link, depth+1)
			}
		}
	})

	return nil
}

// GetCollectedData retrieves the collected data from the Crawler.
func (c *Crawler) GetCollectedData() map[string][]string {
	return c.collectedData.GetData()
}

// SetFilterDomain sets the domain to filter URLs during crawling.
func (c *Crawler) SetFilterDomain(domain string) {
	c.filterDomain = domain
}

// Wait waits for all the crawling tasks to finish.
func (c *Crawler) Wait() {
	c.wg.Wait()
}

// fetch fetches the URL using the HTTP client.
func (c *Crawler) fetch(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers, e.g., User-Agent
	req.Header.Set("User-Agent", "our-crawler-name")

	// Use the client to send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

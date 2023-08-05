// main.go

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"io/ioutil"

	"github.com/Mdromi/golang-search-engine/search-engine/crawler"
	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/Mdromi/golang-search-engine/search-engine/search"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MaxDepth         int    `yaml:"maxDepth"`
	Concurrency      int    `yaml:"concurrency"`
	BoltDBPath       string `yaml:"boltDBPath"`
	RedisAddress     string `yaml:"redisAddress"`
	FilterDomain     string `yaml:"filterDomain"`
	ExampleQueryLink string `yaml:"exampleQueryLink"`
}

func readConfig() (*Config, error) {
	configFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	// Read configuration from config.yaml
	config, err := readConfig()
	if err != nil {
		logrus.Fatal("Failed to read config:", err)
	}

	// Set up logging
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Set up a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a signal channel to handle OS interrupts
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Set up BoltDB
	db, cleanup := indexer.NewInMemoryBoltDB()
	defer func() {
		if err := cleanup(); err != nil {
			logrus.Fatal("Failed to close in-memory BoltDB:", err)
		}
	}()

	// Set up Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	// Initialize the indexer
	idx := indexer.NewIndexer(db, redisClient)

	// Set up the crawler
	c := crawler.NewCrawler(config.MaxDepth, config.Concurrency) // Set maxDepth and concurrency as desired
	c.SetFilterDomain(config.FilterDomain)                       // Set the domain to filter URLs during crawling

	// Start crawling from the provided URL with depth 0 in a separate goroutine
	go func() {
		log.Info("Starting crawling...")
		c.Crawl(config.ExampleQueryLink, 0)
		log.Info("Crawling finished.")
	}()

	// Wait for the OS interrupt signal or a context cancellation
	select {
	case <-signalCh:
		log.Info("Received OS interrupt signal. Gracefully shutting down...")
	case <-ctx.Done():
		log.Info("Context canceled. Gracefully shutting down...")
	}

	// Cancel the context to signal the other goroutines to stop
	cancel()

	// Wait for the crawling to finish and clean up resources
	c.Wait()

	// After crawling, index the collected data
	log.Info("Indexing data...")
	if err := idx.Index(c.GetCollectedData()); err != nil {
		log.Fatal("Failed to index data:", err)
	}
	log.Info("Indexing finished.")

	// Initialize the searcher
	s := search.NewSearcher(db.DB)

	// Query the search engine
	log.Info("Searching...")
	query := "Golang"
	options := &search.SearchOptions{
		FilterDomain: config.FilterDomain,
		SortBy:       "relevance", // or "date"
	}
	results, err := s.Search(query, options)
	if err != nil {
		log.Fatal("Failed to search:", err)
	}
	log.Info("Searching finished.")

	// Print the search results
	log.Info("Search Results:")
	for i, url := range results {
		fmt.Printf("%d. %s\n", i+1, url)
	}
}

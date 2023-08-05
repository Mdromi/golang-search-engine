package main

import (
	"fmt"
	"io/ioutil"

	"github.com/Mdromi/golang-search-engine/search-engine/crawler"
	"github.com/Mdromi/golang-search-engine/search-engine/indexer"
	"github.com/Mdromi/golang-search-engine/search-engine/search"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

/*
ARTICLE: https://voskan.host/2023/08/04/building-search-engine-in-golang/

GOOGLE DOCS: https://docs.google.com/document/d/1NYMRlQsVz6GqAlzDo_XdZpLEEurOWo2EO6fsLrk_q5Y/edit?usp=sharing

TASK: 1.Fixing main.go code 2.Create Docs 3.More Testing 4.Implement Advanced Feature(DOCS)
*/

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

	// Set up BoltDB
	// db, cleanup, err := indexer.NewBoltDB(config.BoltDBPath)
	// if err != nil {
	// 	log.Fatal("Failed to set up BoltDB:", err)
	// }
	// defer func() {
	// 	if err := cleanup(); err != nil {
	// 		log.Fatal("Failed to close BoltDB:", err)
	// 	}
	// }()

	db, cleanup := indexer.NewInMemoryBoltDB()
	defer func() {
		if err := cleanup(); err != nil {
			logrus.Fatal("Failed to close in-memory BoltDB:", err)
		}
	}()

	// Set up Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})
	defer redisClient.Close()

	// Initialize the indexer
	idx := indexer.NewIndexer(db.DB, redisClient)

	// Set up the crawler
	c := crawler.NewCrawler(config.MaxDepth, config.Concurrency)
	c.SetFilterDomain(config.FilterDomain)

	// Start crawling from the provided URL with depth 0
	log.Info("Starting crawling...")
	c.Crawl(config.ExampleQueryLink, 0)
	log.Info("Crawling finished.")

	// Wait for crawling to finish
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
	query := "Phones category"
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
	fmt.Println("results", results)
	for i, url := range results {
		fmt.Printf("%d. %s\n", i+1, url)
	}
}

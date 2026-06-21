package rss

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
)

type FeedConfig struct {
	URL      string `yaml:"url"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type Config struct {
	Feeds []FeedConfig `yaml:"feeds"`
}

type Article struct {
	Title       string
	URL         string
	Description string
	Source      string
	Category    string
	PublishedAt time.Time
	FullText    string
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func FetchAll(cfg *Config) []Article {
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		articles []Article
	)

	for _, feed := range cfg.Feeds {
		wg.Add(1)
		go func(f FeedConfig) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			parser := gofeed.NewParser()
			parsed, err := parser.ParseURLWithContext(f.URL, ctx)
			if err != nil {
				log.Printf("failed to fetch %s: %v", f.Name, err)
				return
			}

			cutoff := time.Now().Add(-24 * time.Hour)

			mu.Lock()
			defer mu.Unlock()

			for _, item := range parsed.Items {
				if item.PublishedParsed == nil {
					continue
				}
				published := *item.PublishedParsed

				if published.Before(cutoff) {
					continue
				}

				desc := item.Description
				if desc == "" {
					desc = item.Content
				}

				articles = append(articles, Article{
					Title:       item.Title,
					URL:         item.Link,
					Description: desc,
					Source:      f.Name,
					Category:    f.Category,
					PublishedAt: published,
				})
			}
		}(feed)
	}

	wg.Wait()
	return articles
}

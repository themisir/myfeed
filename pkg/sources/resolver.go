package sources

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type Item struct {
	Title       string
	Description string
	Url         string
	PublishedAt *time.Time
	UpdatedAt   *time.Time
}

type ResolvedSource struct {
	Title string
	Items []*Item
}

type Resolver interface {
	Resolve(url string) (*ResolvedSource, error)
}

type resolver struct{}

func (r *resolver) Resolve(feedUrl string) (*ResolvedSource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Parse url
	parsedUrl, err := url.Parse(feedUrl)
	if err != nil {
		return nil, err
	}

	// Parse feed
	parser := gofeed.NewParser()
	feed, err := parser.ParseURLWithContext(feedUrl, ctx)
	if err != nil {
		return nil, err
	}

	// Map feed
	source := &ResolvedSource{
		Title: parsedUrl.Hostname(),
		Items: make([]*Item, len(feed.Items)),
	}

	// Map items
	i := 0
	for _, item := range feed.Items {
		itemUrl := item.Link

		if itemUrl == "" {
			// Skip empty posts
			continue
		}

		if !strings.HasPrefix(itemUrl, "https://") && !strings.HasPrefix(itemUrl, "http://") {
			// Support relative URLs
			continue
		}

		source.Items[i] = &Item{
			Url:         itemUrl,
			Title:       item.Title,
			Description: item.Description,
			PublishedAt: item.PublishedParsed,
			UpdatedAt:   item.UpdatedParsed,
		}
		i++
	}

	source.Items = source.Items[:i]

	return source, nil
}

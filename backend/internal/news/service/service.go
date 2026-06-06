// Package service aggregates curated RSS/Atom feeds into a single, cached news
// list. Browsers can't fetch third-party feeds (CORS), so this runs server-side.
package service

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/kinkando/personal-dashboard/internal/news"
	"go.uber.org/zap"
)

const (
	cacheTTL    = 30 * time.Minute
	feedTimeout = 8 * time.Second
	maxItems    = 60
	maxSummary  = 220
)

type Service struct {
	log    *zap.Logger
	parser *gofeed.Parser

	mu        sync.Mutex
	items     []news.Item
	fetchedAt time.Time
}

func New(log *zap.Logger) *Service {
	return &Service{log: log, parser: gofeed.NewParser()}
}

// List returns the cached aggregated feed, refreshing it when stale. Concurrent
// callers are serialised by the mutex so only one refresh runs at a time.
func (s *Service) List(ctx context.Context) ([]news.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if time.Since(s.fetchedAt) < cacheTTL && len(s.items) > 0 {
		return s.items, nil
	}

	items := s.refresh(ctx)
	if len(items) == 0 {
		if len(s.items) > 0 {
			return s.items, nil // all feeds failed — serve stale cache
		}
		return nil, fmt.Errorf("news: no feeds could be fetched")
	}

	s.items = items
	s.fetchedAt = time.Now()
	return s.items, nil
}

// refresh fetches every feed concurrently, skipping any that fail, and returns
// the normalized + sorted + capped item list (newest first, newest featured).
func (s *Service) refresh(ctx context.Context) []news.Item {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		all     []news.Item
		seen    = map[string]bool{}
	)

	for _, f := range news.Feeds() {
		wg.Add(1)
		go func(f news.Feed) {
			defer wg.Done()
			fctx, cancel := context.WithTimeout(ctx, feedTimeout)
			defer cancel()

			parsed, err := s.parser.ParseURLWithContext(f.URL, fctx)
			if err != nil {
				s.log.Warn("news: feed fetch failed", zap.String("url", f.URL), zap.Error(err))
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, it := range parsed.Items {
				url := strings.TrimSpace(it.Link)
				if url == "" || seen[url] {
					continue
				}
				seen[url] = true
				all = append(all, news.Item{
					ID:          itemID(it),
					Title:       strings.TrimSpace(it.Title),
					Summary:     summarize(it),
					Category:    f.Category,
					Source:      f.Source,
					URL:         url,
					PublishedAt: publishedAt(it),
				})
			}
		}(f)
	}
	wg.Wait()

	sort.Slice(all, func(i, j int) bool {
		return all[i].PublishedAt.After(all[j].PublishedAt)
	})
	if len(all) > maxItems {
		all = all[:maxItems]
	}
	if len(all) > 0 {
		all[0].Featured = true
	}
	return all
}

// ── Helpers ─────────────────────────────────────────────────────────────────

var tagRE = regexp.MustCompile(`<[^>]*>`)
var wsRE = regexp.MustCompile(`\s+`)

// summarize derives a clean plain-text summary from the item's description or
// content: strip HTML, unescape entities, collapse whitespace, truncate.
func summarize(it *gofeed.Item) string {
	raw := it.Description
	if strings.TrimSpace(raw) == "" {
		raw = it.Content
	}
	text := html.UnescapeString(tagRE.ReplaceAllString(raw, " "))
	text = strings.TrimSpace(wsRE.ReplaceAllString(text, " "))
	r := []rune(text)
	if len(r) > maxSummary {
		return strings.TrimSpace(string(r[:maxSummary])) + "…"
	}
	return text
}

func publishedAt(it *gofeed.Item) time.Time {
	if it.PublishedParsed != nil {
		return *it.PublishedParsed
	}
	if it.UpdatedParsed != nil {
		return *it.UpdatedParsed
	}
	return time.Now()
}

func itemID(it *gofeed.Item) string {
	if strings.TrimSpace(it.GUID) != "" {
		return it.GUID
	}
	return it.Link
}

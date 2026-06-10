package service

import (
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

// ── summarize ─────────────────────────────────────────────────────────────────

func TestSummarize_PlainText(t *testing.T) {
	it := &gofeed.Item{Description: "Hello World"}
	got := summarize(it)
	if got != "Hello World" {
		t.Errorf("summarize() = %q, want %q", got, "Hello World")
	}
}

func TestSummarize_StripsHTMLTags(t *testing.T) {
	it := &gofeed.Item{Description: "<p>Hello <b>World</b></p>"}
	got := summarize(it)
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("summarize() = %q, still contains HTML tags", got)
	}
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "World") {
		t.Errorf("summarize() = %q, lost text content", got)
	}
}

func TestSummarize_UnescapesHTMLEntities(t *testing.T) {
	it := &gofeed.Item{Description: "AT&amp;T &lt;3"}
	got := summarize(it)
	if !strings.Contains(got, "AT&T") {
		t.Errorf("summarize() = %q, expected & to be unescaped", got)
	}
}

func TestSummarize_CollapsesWhitespace(t *testing.T) {
	it := &gofeed.Item{Description: "  foo   bar\t\nbaz  "}
	got := summarize(it)
	if got != "foo bar baz" {
		t.Errorf("summarize() = %q, want %q", got, "foo bar baz")
	}
}

func TestSummarize_TruncatesAt220Runes(t *testing.T) {
	// Build a 300-rune string of Thai characters (multi-byte but single rune each).
	long := strings.Repeat("ก", 300)
	it := &gofeed.Item{Description: long}
	got := summarize(it)
	r := []rune(got)
	// Should end with "…" and be no longer than 221 runes (220 text + ellipsis).
	if !strings.HasSuffix(got, "…") {
		t.Errorf("summarize() = %q, want trailing ellipsis", got)
	}
	if len(r) > 221 {
		t.Errorf("summarize() rune length = %d, want ≤ 221", len(r))
	}
}

func TestSummarize_Exactly220RunesNoEllipsis(t *testing.T) {
	exact := strings.Repeat("a", maxSummary) // exactly 220 'a's
	it := &gofeed.Item{Description: exact}
	got := summarize(it)
	if strings.HasSuffix(got, "…") {
		t.Errorf("summarize() added ellipsis for exactly %d runes: %q", maxSummary, got)
	}
}

func TestSummarize_FallsBackToContent(t *testing.T) {
	// Description is blank — should fall back to Content.
	it := &gofeed.Item{Description: "   ", Content: "Fallback content"}
	got := summarize(it)
	if !strings.Contains(got, "Fallback content") {
		t.Errorf("summarize() = %q, expected content fallback", got)
	}
}

// ── publishedAt ───────────────────────────────────────────────────────────────

func TestPublishedAt_UsesPublishedParsed(t *testing.T) {
	ts := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	it := &gofeed.Item{PublishedParsed: &ts, UpdatedParsed: nil}
	got := publishedAt(it)
	if !got.Equal(ts) {
		t.Errorf("publishedAt() = %v, want %v", got, ts)
	}
}

func TestPublishedAt_FallsBackToUpdatedParsed(t *testing.T) {
	ts := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	it := &gofeed.Item{PublishedParsed: nil, UpdatedParsed: &ts}
	got := publishedAt(it)
	if !got.Equal(ts) {
		t.Errorf("publishedAt() = %v, want %v", got, ts)
	}
}

func TestPublishedAt_FallsBackToNowWhenBothNil(t *testing.T) {
	before := time.Now()
	got := publishedAt(&gofeed.Item{})
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Errorf("publishedAt() fallback = %v, expected between %v and %v", got, before, after)
	}
}

// ── itemID ────────────────────────────────────────────────────────────────────

func TestItemID_UsesGUIDWhenSet(t *testing.T) {
	it := &gofeed.Item{GUID: "guid-123", Link: "https://example.com/article"}
	got := itemID(it)
	if got != "guid-123" {
		t.Errorf("itemID() = %q, want %q", got, "guid-123")
	}
}

func TestItemID_FallsBackToLinkWhenGUIDEmpty(t *testing.T) {
	it := &gofeed.Item{GUID: "", Link: "https://example.com/article"}
	got := itemID(it)
	if got != "https://example.com/article" {
		t.Errorf("itemID() = %q, want %q", got, "https://example.com/article")
	}
}

func TestItemID_FallsBackToLinkWhenGUIDWhitespace(t *testing.T) {
	it := &gofeed.Item{GUID: "   ", Link: "https://example.com/article"}
	got := itemID(it)
	if got != "https://example.com/article" {
		t.Errorf("itemID() = %q, want link when GUID is whitespace", got)
	}
}


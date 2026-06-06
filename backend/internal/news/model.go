package news

import "time"

// Category mirrors the frontend NewsCategory union.
type Category string

const (
	CategoryAI       Category = "ai"
	CategoryWebDev   Category = "webdev"
	CategoryCloud    Category = "cloud"
	CategorySecurity Category = "security"
	CategoryGameTech Category = "gametech"
)

// Item is a single normalized news article served to the frontend.
type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Category    Category  `json:"category"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	Featured    bool      `json:"featured"`
}

// Feed is one curated RSS/Atom source. Category and Source are fixed per feed so
// the parsed items inherit a consistent label regardless of feed metadata.
type Feed struct {
	URL      string
	Source   string
	Category Category
}

// Feeds returns the curated registry (read-only).
func Feeds() []Feed { return feeds }

// feeds is the curated registry. URLs are tunable; each entry fixes the item's
// Category and display Source.
var feeds = []Feed{
	// AI
	{URL: "https://hnrss.org/newest?q=AI+OR+LLM&points=50", Source: "Hacker News", Category: CategoryAI},
	{URL: "https://www.artificialintelligence-news.com/feed/", Source: "AI News", Category: CategoryAI},

	// Web Dev
	{URL: "https://dev.to/feed", Source: "DEV", Category: CategoryWebDev},
	{URL: "https://css-tricks.com/feed/", Source: "CSS-Tricks", Category: CategoryWebDev},

	// Cloud
	{URL: "https://aws.amazon.com/blogs/aws/feed/", Source: "AWS News", Category: CategoryCloud},
	{URL: "https://www.infoq.com/cloud-computing/rss/", Source: "InfoQ Cloud", Category: CategoryCloud},

	// Security
	{URL: "https://feeds.feedburner.com/TheHackersNews", Source: "The Hacker News", Category: CategorySecurity},
	{URL: "https://www.bleepingcomputer.com/feed/", Source: "BleepingComputer", Category: CategorySecurity},

	// Game Tech
	{URL: "https://www.gamedeveloper.com/rss.xml", Source: "Game Developer", Category: CategoryGameTech},
}

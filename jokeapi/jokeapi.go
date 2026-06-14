// Package jokeapi is the library behind the jokeapi command line:
// the HTTP client, request shaping, and the typed data models for v2.jokeapi.dev.
//
// The Client sets a real User-Agent, paces requests so a busy session stays
// polite, and retries the transient failures (429 and 5xx) that any public
// site throws under load.
package jokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "v2.jokeapi.dev"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://v2.jokeapi.dev",
		UserAgent: "jokeapi-cli/0.1.0 (github.com/tamnd/jokeapi-cli)",
		Rate:      200 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to v2.jokeapi.dev over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Jokes fetches count jokes from the given category.
// If safe is true, only family-safe jokes are returned.
// Category defaults to "Any" if empty.
// blacklist is a comma-separated list of flags to exclude (nsfw, religious, political, racist, sexist, explicit).
func (c *Client) Jokes(ctx context.Context, category string, safe bool, count int, blacklist string) ([]Joke, error) {
	if category == "" {
		category = "Any"
	}
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}
	u := fmt.Sprintf("%s/joke/%s?amount=%d", c.cfg.BaseURL, category, count)
	if safe {
		u += "&safe-mode"
	}
	if blacklist != "" {
		u += "&blacklistFlags=" + strings.TrimSpace(blacklist)
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	// The API returns a flat object when amount=1, and a jokes array when amount>=2.
	// We always try the array form first; fall back to single-joke form.
	var resp jokesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode jokes: %w", err)
	}

	var raws []rawJoke
	if len(resp.Jokes) > 0 {
		raws = resp.Jokes
	} else {
		// Single joke: the whole body is a rawJoke.
		var single rawJoke
		if err := json.Unmarshal(body, &single); err != nil {
			return nil, fmt.Errorf("decode single joke: %w", err)
		}
		raws = []rawJoke{single}
	}

	jokes := make([]Joke, 0, len(raws))
	for i, r := range raws {
		j := Joke{
			Rank:     i + 1,
			ID:       r.ID,
			Category: r.Category,
			Type:     r.Type,
			Safe:     r.Safe,
		}
		if r.Type == "twopart" {
			j.Setup = r.Setup
			j.Delivery = r.Delivery
		} else {
			j.Text = r.Joke
		}
		jokes = append(jokes, j)
	}
	return jokes, nil
}

// Categories returns the list of available joke categories from the API.
func (c *Client) Categories(ctx context.Context) ([]string, error) {
	u := c.cfg.BaseURL + "/categories"
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp categoriesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode categories: %w", err)
	}
	return resp.Categories, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}

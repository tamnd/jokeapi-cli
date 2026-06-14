// Package jokeapi is the library behind the jokeapi command line:
// the HTTP client, request shaping, and the typed data models for
// official-joke-api.appspot.com.
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
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "official-joke-api.appspot.com"

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
		BaseURL:   "https://official-joke-api.appspot.com",
		UserAgent: "jokeapi-cli/0.1 (tamnd87@gmail.com)",
		Rate:      200 * time.Millisecond,
		Timeout:   10 * time.Second,
		Retries:   3,
	}
}

// Client talks to official-joke-api.appspot.com over HTTP.
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

// wireJoke is the JSON shape returned by the API.
type wireJoke struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

func wireToJoke(w wireJoke) Joke {
	return Joke(w)
}

// Random fetches count random jokes. If count == 1, it uses /random_joke
// (returns a single object). If count > 1, it uses /jokes/random/{count}
// (returns an array).
func (c *Client) Random(ctx context.Context, count int) ([]Joke, error) {
	if count <= 1 {
		body, err := c.get(ctx, c.cfg.BaseURL+"/random_joke")
		if err != nil {
			return nil, err
		}
		var w wireJoke
		if err := json.Unmarshal(body, &w); err != nil {
			return nil, fmt.Errorf("decode joke: %w", err)
		}
		return []Joke{wireToJoke(w)}, nil
	}

	u := fmt.Sprintf("%s/jokes/random/%d", c.cfg.BaseURL, count)
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var ws []wireJoke
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, fmt.Errorf("decode jokes: %w", err)
	}
	jokes := make([]Joke, len(ws))
	for i, w := range ws {
		jokes[i] = wireToJoke(w)
	}
	return jokes, nil
}

// ByType fetches jokes of a given type. If count <= 1, it uses
// /jokes/{type}/random (returns array of 1). If count > 1 (up to 10), it uses
// /jokes/{type}/ten (returns up to 10 jokes, trimmed to count).
func (c *Client) ByType(ctx context.Context, jokeType string, count int) ([]Joke, error) {
	var u string
	if count <= 1 {
		u = fmt.Sprintf("%s/jokes/%s/random", c.cfg.BaseURL, jokeType)
	} else {
		u = fmt.Sprintf("%s/jokes/%s/ten", c.cfg.BaseURL, jokeType)
	}

	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var ws []wireJoke
	if err := json.Unmarshal(body, &ws); err != nil {
		return nil, fmt.Errorf("decode jokes: %w", err)
	}
	if count > 0 && len(ws) > count {
		ws = ws[:count]
	}
	jokes := make([]Joke, len(ws))
	for i, w := range ws {
		jokes[i] = wireToJoke(w)
	}
	return jokes, nil
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

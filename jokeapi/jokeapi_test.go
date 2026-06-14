package jokeapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/jokeapi-cli/jokeapi"
)

const fakeRandomJokeJSON = `{"type":"general","setup":"What do you get hanging from Apple trees?","punchline":"Pears.","id":1}`

const fakeRandomJokesJSON = `[{"type":"programming","setup":"Why do Java developers wear glasses?","punchline":"Because they don't C#","id":8},{"type":"general","setup":"Why do cows wear bells?","punchline":"Because their horns don't work.","id":2}]`

const fakeTypeRandomJSON = `[{"type":"programming","setup":"Why do Java developers wear glasses?","punchline":"Because they don't C#","id":8}]`

const fakeTypeTenJSON = `[{"type":"programming","setup":"Why do Java developers wear glasses?","punchline":"Because they don't C#","id":8},{"type":"programming","setup":"How many programmers does it take to change a light bulb?","punchline":"None, that's a hardware problem.","id":23},{"type":"programming","setup":"What's the object-oriented way to become wealthy?","punchline":"Inheritance.","id":37}]`

func newTestClient(ts *httptest.Server) *jokeapi.Client {
	cfg := jokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return jokeapi.NewClient(cfg)
}

// --- Random tests ---

func TestRandomSingleSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeRandomJokeJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Random(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestRandomSingleUsesRandomJokeEndpoint(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeRandomJokeJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.Random(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/random_joke" {
		t.Errorf("path = %q, want /random_joke", gotPath)
	}
	if len(jokes) != 1 {
		t.Fatalf("len = %d, want 1", len(jokes))
	}
	if jokes[0].ID != 1 {
		t.Errorf("ID = %d, want 1", jokes[0].ID)
	}
	if jokes[0].Setup != "What do you get hanging from Apple trees?" {
		t.Errorf("Setup = %q, unexpected", jokes[0].Setup)
	}
	if jokes[0].Punchline != "Pears." {
		t.Errorf("Punchline = %q, want Pears.", jokes[0].Punchline)
	}
}

func TestRandomMultipleUsesRandomCountEndpoint(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeRandomJokesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.Random(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/jokes/random/2" {
		t.Errorf("path = %q, want /jokes/random/2", gotPath)
	}
	if len(jokes) != 2 {
		t.Fatalf("len = %d, want 2", len(jokes))
	}
}

func TestRandomMultipleParsesFields(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeRandomJokesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.Random(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if jokes[0].Type != "programming" {
		t.Errorf("jokes[0].Type = %q, want programming", jokes[0].Type)
	}
	if jokes[0].ID != 8 {
		t.Errorf("jokes[0].ID = %d, want 8", jokes[0].ID)
	}
	if jokes[1].Type != "general" {
		t.Errorf("jokes[1].Type = %q, want general", jokes[1].Type)
	}
}

// --- ByType tests ---

func TestByTypeSingleUsesRandomEndpoint(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeTypeRandomJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.ByType(context.Background(), "programming", 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/jokes/programming/random" {
		t.Errorf("path = %q, want /jokes/programming/random", gotPath)
	}
	if len(jokes) != 1 {
		t.Fatalf("len = %d, want 1", len(jokes))
	}
	if jokes[0].Type != "programming" {
		t.Errorf("Type = %q, want programming", jokes[0].Type)
	}
}

func TestByTypeMultipleUsesTenEndpoint(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeTypeTenJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.ByType(context.Background(), "programming", 3)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/jokes/programming/ten" {
		t.Errorf("path = %q, want /jokes/programming/ten", gotPath)
	}
	if len(jokes) != 3 {
		t.Fatalf("len = %d, want 3", len(jokes))
	}
}

func TestByTypeMultipleTrimmedToCount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeTypeTenJSON) // returns 3 jokes
	}))
	defer ts.Close()

	c := newTestClient(ts)
	jokes, err := c.ByType(context.Background(), "programming", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(jokes) != 2 {
		t.Fatalf("len = %d, want 2 (trimmed)", len(jokes))
	}
}

func TestByTypeSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeTypeRandomJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.ByType(context.Background(), "general", 1)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestByTypeTypeInPath(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeTypeRandomJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.ByType(context.Background(), "knock-knock", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, "knock-knock") {
		t.Errorf("path %q does not contain knock-knock", gotPath)
	}
}

// --- Retry test ---

func TestRandomRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeRandomJokeJSON)
	}))
	defer ts.Close()

	cfg := jokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := jokeapi.NewClient(cfg)

	_, err := c.Random(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

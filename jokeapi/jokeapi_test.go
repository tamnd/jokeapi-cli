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

const fakeSingleJSON = `{"error":false,"category":"Misc","type":"single","joke":"Schrödinger's cat walks into a bar and doesn't.","flags":{"nsfw":false,"religious":false,"political":false,"racist":false,"sexist":false,"explicit":false},"id":209,"safe":true,"lang":"en"}`

const fakeTwopartJSON = `{"error":false,"category":"Programming","type":"twopart","setup":"How did the programmer die in the shower?","delivery":"He read the shampoo bottle instructions: Lather. Rinse. Repeat.","flags":{"nsfw":false,"religious":false,"political":false,"racist":false,"sexist":false,"explicit":false},"id":9,"safe":true,"lang":"en"}`

const fakeMockJSON = `{"error":false,"amount":2,"jokes":[{"category":"Programming","type":"twopart","setup":"Why do Java developers wear glasses?","delivery":"Because they don't C#","flags":{"nsfw":false,"religious":false,"political":false,"racist":false,"sexist":false,"explicit":false},"safe":true,"id":8,"lang":"en"},{"category":"Misc","type":"single","joke":"I've got a really good joke about construction, but I'm still working on it.","flags":{"nsfw":false,"religious":false,"political":false,"racist":false,"sexist":false,"explicit":false},"safe":true,"id":15,"lang":"en"}]}`

const fakeCategoriesJSON = `{"error":false,"categories":["Any","Misc","Programming","Dark","Pun","Spooky","Christmas"],"categoryAliases":[],"timestamp":1718000000}`

func newTestClient(ts *httptest.Server) *jokeapi.Client {
	cfg := jokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return jokeapi.NewClient(cfg)
}

// --- Joke (single) tests ---

func TestJokeSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Any", "", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestJokeParsesSingle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	j, err := c.Joke(context.Background(), "Misc", "single", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if j.Type != "single" {
		t.Errorf("Type = %q, want single", j.Type)
	}
	if j.Joke == "" {
		t.Error("Joke field empty for single type")
	}
	if j.Setup != "" || j.Delivery != "" {
		t.Error("Setup/Delivery should be empty for single type")
	}
	if j.ID != 209 {
		t.Errorf("ID = %d, want 209", j.ID)
	}
	if j.Lang != "en" {
		t.Errorf("Lang = %q, want en", j.Lang)
	}
}

func TestJokeParsesTwopart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeTwopartJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	j, err := c.Joke(context.Background(), "Programming", "twopart", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if j.Type != "twopart" {
		t.Errorf("Type = %q, want twopart", j.Type)
	}
	if j.Setup == "" {
		t.Error("Setup empty for twopart joke")
	}
	if j.Delivery == "" {
		t.Error("Delivery empty for twopart joke")
	}
	if j.Joke != "" {
		t.Errorf("Joke should be empty for twopart, got %q", j.Joke)
	}
}

func TestJokeSafeMode(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Any", "", "en", true, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "safe-mode") {
		t.Errorf("URL %q does not contain safe-mode", gotURL)
	}
}

func TestJokeBlacklist(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Any", "", "en", false, "nsfw,racist")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "blacklistFlags") {
		t.Errorf("URL %q does not contain blacklistFlags", gotURL)
	}
	if !strings.Contains(gotURL, "nsfw") {
		t.Errorf("URL %q does not contain nsfw flag", gotURL)
	}
}

func TestJokeLangInURL(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Any", "", "de", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "lang=de") {
		t.Errorf("URL %q does not contain lang=de", gotURL)
	}
}

func TestJokeTypeInURL(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Any", "single", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "type=single") {
		t.Errorf("URL %q does not contain type=single", gotURL)
	}
}

func TestJokeCategoryInURL(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeSingleJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Joke(context.Background(), "Programming", "", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, "Programming") {
		t.Errorf("URL path %q does not contain Programming", gotPath)
	}
}

// --- Jokes (multiple) tests ---

func TestJokesSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Jokes(context.Background(), "Any", false, 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestJokesParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Jokes(context.Background(), "Any", false, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Type != "twopart" {
		t.Errorf("items[0].Type = %q, want twopart", items[0].Type)
	}
	if items[0].Setup != "Why do Java developers wear glasses?" {
		t.Errorf("items[0].Setup = %q, unexpected", items[0].Setup)
	}
	if items[0].Delivery != "Because they don't C#" {
		t.Errorf("items[0].Delivery = %q, unexpected", items[0].Delivery)
	}
	if items[0].Joke != "" {
		t.Errorf("items[0].Joke = %q, want empty for twopart", items[0].Joke)
	}
	if items[1].Type != "single" {
		t.Errorf("items[1].Type = %q, want single", items[1].Type)
	}
	if items[1].Joke == "" {
		t.Error("items[1].Joke is empty, want non-empty for single joke")
	}
	if items[1].Setup != "" {
		t.Errorf("items[1].Setup = %q, want empty for single", items[1].Setup)
	}
	if items[1].Delivery != "" {
		t.Errorf("items[1].Delivery = %q, want empty for single", items[1].Delivery)
	}
}

func TestJokesCategoryInURL(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Jokes(context.Background(), "Programming", false, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, "Programming") {
		t.Errorf("URL path %q does not contain Programming", gotPath)
	}
}

func TestJokesSafeMode(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Jokes(context.Background(), "Any", true, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "safe-mode") {
		t.Errorf("URL %q does not contain safe-mode", gotURL)
	}
}

func TestJokesRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	cfg := jokeapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := jokeapi.NewClient(cfg)

	_, err := c.Jokes(context.Background(), "Any", false, 2)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

// --- Categories tests ---

func TestCategories_basic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/categories") {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) == 0 {
		t.Error("categories list is empty, want non-empty")
	}
}

func TestCategories_parsesAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"Any": true, "Misc": true, "Programming": true, "Dark": true,
		"Pun": true, "Spooky": true, "Christmas": true,
	}
	for _, cat := range cats {
		delete(want, cat.Name)
	}
	if len(want) > 0 {
		t.Errorf("missing categories: %v", want)
	}
}

func TestCategories_returnsCategoryStruct(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) == 0 {
		t.Fatal("no categories returned")
	}
	if cats[0].Name == "" {
		t.Error("Category.Name is empty")
	}
}

func TestJokeFlags_parsed(t *testing.T) {
	nsfwJSON := `{"error":false,"category":"Dark","type":"single","joke":"dark joke","flags":{"nsfw":true,"religious":false,"political":false,"racist":false,"sexist":false,"explicit":false},"id":100,"safe":false,"lang":"en"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, nsfwJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	j, err := c.Joke(context.Background(), "Dark", "", "en", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if !j.NSFW {
		t.Error("NSFW flag should be true")
	}
	if j.Safe {
		t.Error("Safe should be false for nsfw joke")
	}
}

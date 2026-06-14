package jokeapi

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// HTTP behaviour is covered in jokeapi_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "jokeapi" {
		t.Errorf("Scheme = %q, want jokeapi", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "jokeapi" {
		t.Errorf("Identity.Binary = %q, want jokeapi", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"programming", "type", "programming"},
		{"general", "type", "general"},
		{"knock-knock", "type", "knock-knock"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("type", "programming")
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	want := "https://official-joke-api.appspot.com/jokes/programming/random"
	if got != want {
		t.Errorf("Locate = %q, want %q", got, want)
	}
}

func TestLocateGeneral(t *testing.T) {
	got, err := Domain{}.Locate("type", "general")
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	if got == "" {
		t.Error("Locate returned empty URL")
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestWireToJoke(t *testing.T) {
	w := wireJoke{
		ID:        1,
		Type:      "general",
		Setup:     "What do you get hanging from Apple trees?",
		Punchline: "Pears.",
	}
	j := wireToJoke(w)
	if j.ID != 1 {
		t.Errorf("ID = %d, want 1", j.ID)
	}
	if j.Type != "general" {
		t.Errorf("Type = %q, want general", j.Type)
	}
	if j.Setup != "What do you get hanging from Apple trees?" {
		t.Errorf("Setup = %q, unexpected", j.Setup)
	}
	if j.Punchline != "Pears." {
		t.Errorf("Punchline = %q, want Pears.", j.Punchline)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BaseURL != "https://official-joke-api.appspot.com" {
		t.Errorf("BaseURL = %q, want https://official-joke-api.appspot.com", cfg.BaseURL)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
	if cfg.Retries <= 0 {
		t.Errorf("Retries = %d, want > 0", cfg.Retries)
	}
}

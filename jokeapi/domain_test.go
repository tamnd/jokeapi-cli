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
		{"Programming", "category", "Programming"},
		{"dark", "category", "dark"},
		{"Any", "category", "Any"},
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
	got, err := Domain{}.Locate("category", "Programming")
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	want := "https://v2.jokeapi.dev/joke/Programming"
	if got != want {
		t.Errorf("Locate = %q, want %q", got, want)
	}
}

func TestLocateAny(t *testing.T) {
	got, err := Domain{}.Locate("category", "Any")
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

func TestRawToJokeSingle(t *testing.T) {
	r := rawJoke{
		ID:       42,
		Category: "Misc",
		Type:     "single",
		Joke:     "Why did the chicken cross the road?",
		Safe:     true,
		Lang:     "en",
		Flags: rawFlags{
			NSFW:   false,
			Racist: false,
		},
	}
	j := rawToJoke(r)
	if j.ID != 42 {
		t.Errorf("ID = %d, want 42", j.ID)
	}
	if j.Joke != "Why did the chicken cross the road?" {
		t.Errorf("Joke = %q, unexpected", j.Joke)
	}
	if j.Setup != "" || j.Delivery != "" {
		t.Errorf("Setup/Delivery should be empty for single type")
	}
	if j.Lang != "en" {
		t.Errorf("Lang = %q, want en", j.Lang)
	}
}

func TestRawToJokeTwoPart(t *testing.T) {
	r := rawJoke{
		ID:       9,
		Category: "Programming",
		Type:     "twopart",
		Setup:    "How did the programmer die in the shower?",
		Delivery: "He read the shampoo bottle instructions: Lather. Rinse. Repeat.",
		Safe:     true,
		Lang:     "en",
		Flags: rawFlags{
			NSFW:   false,
			Sexist: false,
		},
	}
	j := rawToJoke(r)
	if j.Setup != "How did the programmer die in the shower?" {
		t.Errorf("Setup = %q, unexpected", j.Setup)
	}
	if j.Delivery != "He read the shampoo bottle instructions: Lather. Rinse. Repeat." {
		t.Errorf("Delivery = %q, unexpected", j.Delivery)
	}
	if j.Joke != "" {
		t.Errorf("Joke should be empty for twopart type, got %q", j.Joke)
	}
}

func TestRawToJokeFlags(t *testing.T) {
	r := rawJoke{
		ID:   99,
		Type: "single",
		Joke: "test joke",
		Flags: rawFlags{
			NSFW:      true,
			Religious: true,
			Political: false,
			Racist:    true,
			Sexist:    false,
			Explicit:  true,
		},
	}
	j := rawToJoke(r)
	if !j.NSFW {
		t.Error("NSFW should be true")
	}
	if !j.Religious {
		t.Error("Religious should be true")
	}
	if j.Political {
		t.Error("Political should be false")
	}
	if !j.Racist {
		t.Error("Racist should be true")
	}
	if j.Sexist {
		t.Error("Sexist should be false")
	}
	if !j.Explicit {
		t.Error("Explicit should be true")
	}
}

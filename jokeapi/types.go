package jokeapi

// Joke is one joke from official-joke-api.appspot.com.
type Joke struct {
	ID        int    `kit:"id" json:"id"`
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
}

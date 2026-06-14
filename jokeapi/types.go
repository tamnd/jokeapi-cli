package jokeapi

// Joke is one joke from v2.jokeapi.dev.
type Joke struct {
	Rank     int    `json:"rank"`
	ID       int    `json:"id"`
	Category string `json:"category"`
	Type     string `json:"type"`     // "single" or "twopart"
	Text     string `json:"text"`     // for single jokes
	Setup    string `json:"setup"`    // for twopart jokes
	Delivery string `json:"delivery"` // for twopart jokes
	Safe     bool   `json:"safe"`
}

// rawJoke is the wire shape returned by v2.jokeapi.dev.
type rawJoke struct {
	Error    bool   `json:"error"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Joke     string `json:"joke"`     // set when type=single
	Setup    string `json:"setup"`    // set when type=twopart
	Delivery string `json:"delivery"` // set when type=twopart
	Safe     bool   `json:"safe"`
	ID       int    `json:"id"`
}

// jokesResponse is the top-level JSON envelope for multi-joke responses.
type jokesResponse struct {
	Error  bool      `json:"error"`
	Amount int       `json:"amount"`
	Jokes  []rawJoke `json:"jokes"`
}

// categoriesResponse is the response from GET /categories.
type categoriesResponse struct {
	Error      bool     `json:"error"`
	Categories []string `json:"categories"`
}

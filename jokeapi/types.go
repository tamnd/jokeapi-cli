package jokeapi

// Joke is one joke from v2.jokeapi.dev.
type Joke struct {
	ID        int    `kit:"id" json:"id"`
	Category  string `json:"category"`
	Type      string `json:"type"`     // "single" or "twopart"
	Joke      string `json:"joke"`     // for single-type
	Setup     string `json:"setup"`    // for twopart-type
	Delivery  string `json:"delivery"` // for twopart-type
	Safe      bool   `json:"safe"`
	Lang      string `json:"lang"`
	NSFW      bool   `json:"nsfw"`
	Religious bool   `json:"religious"`
	Political bool   `json:"political"`
	Racist    bool   `json:"racist"`
	Sexist    bool   `json:"sexist"`
	Explicit  bool   `json:"explicit"`
}

// Category is a joke category from v2.jokeapi.dev.
type Category struct {
	Name string `kit:"id" json:"name"`
}

// rawFlags is the wire shape of the flags object.
type rawFlags struct {
	NSFW      bool `json:"nsfw"`
	Religious bool `json:"religious"`
	Political bool `json:"political"`
	Racist    bool `json:"racist"`
	Sexist    bool `json:"sexist"`
	Explicit  bool `json:"explicit"`
}

// rawJoke is the wire shape returned by v2.jokeapi.dev.
type rawJoke struct {
	Error    bool     `json:"error"`
	Category string   `json:"category"`
	Type     string   `json:"type"`
	Joke     string   `json:"joke"`     // set when type=single
	Setup    string   `json:"setup"`    // set when type=twopart
	Delivery string   `json:"delivery"` // set when type=twopart
	Flags    rawFlags `json:"flags"`
	Safe     bool     `json:"safe"`
	Lang     string   `json:"lang"`
	ID       int      `json:"id"`
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

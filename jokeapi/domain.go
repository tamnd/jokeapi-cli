package jokeapi

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes jokeapi as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/jokeapi-cli/jokeapi"
//
// The same Domain also builds the standalone jokeapi binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the jokeapi driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "jokeapi",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "jokeapi",
			Short:  "Fetch categorized jokes from JokeAPI v2.",
			Long: `jokeapi fetches jokes from the free JokeAPI v2 service (v2.jokeapi.dev).
No authentication required.

Categories: Any, Misc, Programming, Dark, Pun, Spooky, Christmas.
Joke types: single (one text field) or twopart (setup + delivery).

Use --safe to restrict to family-safe jokes, --blacklist to exclude
jokes with specific flags (nsfw, religious, political, racist, sexist, explicit).`,
			Site: Host,
			Repo: "https://github.com/tamnd/jokeapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "joke",
		Group:   "read",
		List:    true,
		Summary: "Fetch jokes from JokeAPI v2",
		Args:    []kit.Arg{{Name: "category", Help: "joke category (Any, Misc, Programming, Dark, Pun, Spooky, Christmas)", Optional: true}},
	}, jokeOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "categories",
		Group:   "read",
		Single:  true,
		Summary: "List available joke categories",
	}, categoriesOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type jokeInput struct {
	Category  string  `kit:"arg,optional" help:"category: Any, Misc, Programming, Dark, Pun, Spooky, Christmas"`
	Amount    int     `kit:"flag" help:"number of jokes (1-10, default 1)"`
	Safe      bool    `kit:"flag" help:"restrict to safe jokes only"`
	Blacklist string  `kit:"flag" help:"comma-separated flags to exclude: nsfw,religious,political,racist,sexist,explicit"`
	Client    *Client `kit:"inject"`
}

type categoriesInput struct {
	Client *Client `kit:"inject"`
}

// categoriesResult wraps the list of categories for JSON output.
type categoriesResult struct {
	Categories []string `json:"categories"`
}

// --- handlers ---

func jokeOp(ctx context.Context, in jokeInput, emit func(Joke) error) error {
	amount := in.Amount
	if amount <= 0 {
		amount = 1
	}
	if amount > 10 {
		amount = 10
	}
	cat := in.Category
	if cat == "" {
		cat = "Any"
	}
	items, err := in.Client.Jokes(ctx, cat, in.Safe, amount, in.Blacklist)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func categoriesOp(ctx context.Context, in categoriesInput, emit func(*categoriesResult) error) error {
	cats, err := in.Client.Categories(ctx)
	if err != nil {
		return err
	}
	return emit(&categoriesResult{Categories: cats})
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty jokeapi reference")
	}
	return "jokes", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "jokes":
		return "https://v2.jokeapi.dev/joke/Any", nil
	default:
		return "", errs.Usage("jokeapi has no resource type %q", uriType)
	}
}

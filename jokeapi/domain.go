package jokeapi

import (
	"context"
	"fmt"

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
			Short:  "Fetch jokes from the Official Joke API.",
			Long: `jokeapi fetches jokes from the free Official Joke API
(official-joke-api.appspot.com). No authentication required.

Available types: general, programming, knock-knock, dark, pun, spooky, christmas.`,
			Site: Host,
			Repo: "https://github.com/tamnd/jokeapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "random",
		Group:   "read",
		List:    true,
		Summary: "Get random jokes",
	}, randomOp)

	kit.Handle(app, kit.OpMeta{
		Name:    "type",
		Group:   "read",
		List:    true,
		Summary: "Get jokes by type",
		Args:    []kit.Arg{{Name: "jokeType", Help: "joke type: general|programming|knock-knock|dark|pun|spooky|christmas"}},
	}, typeOp)
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

type randomInput struct {
	Count  int     `kit:"flag" help:"number of jokes to fetch" default:"1"`
	Client *Client `kit:"inject"`
}

type typeInput struct {
	JokeType string  `kit:"arg" help:"joke type: general|programming|knock-knock|dark|pun|spooky|christmas"`
	Count    int     `kit:"flag" help:"number of jokes to fetch (max 10)" default:"1"`
	Client   *Client `kit:"inject"`
}

// --- handlers ---

func randomOp(ctx context.Context, in randomInput, emit func(Joke) error) error {
	count := in.Count
	if count <= 0 {
		count = 1
	}
	jokes, err := in.Client.Random(ctx, count)
	if err != nil {
		return err
	}
	for _, j := range jokes {
		if err := emit(j); err != nil {
			return err
		}
	}
	return nil
}

func typeOp(ctx context.Context, in typeInput, emit func(Joke) error) error {
	count := in.Count
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}
	jokes, err := in.Client.ByType(ctx, in.JokeType, count)
	if err != nil {
		return err
	}
	for _, j := range jokes {
		if err := emit(j); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty jokeapi reference")
	}
	return "type", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "type":
		return fmt.Sprintf("https://official-joke-api.appspot.com/jokes/%s/random", id), nil
	default:
		return "", errs.Usage("jokeapi has no resource type %q", uriType)
	}
}

package checker

import (
	"context"
	"math/rand/v2"
	"time"

	"git.happydns.org/happyDomain/model"
	sdk "git.happydns.org/happyDomain/sdk/checker"
)

// Collect gathers observation data. This is called by happyDomain (or the
// /collect HTTP endpoint) every time a check runs.
//
// In a real checker, this is where you would perform the actual monitoring
// work: sending network requests, querying APIs, measuring latency, etc.
//
// This dummy implementation simply reads options and generates a random score
// so you can focus on the structure rather than external dependencies.
//
// Parameters:
//   - ctx: a context for cancellation/timeout - always honour it.
//   - opts: the merged checker options (admin + user + domain + service + run).
//
// Return:
//   - any: the observation data (will be JSON-serialised by the SDK).
//   - error: non-nil if collection failed entirely.
func (p *dummyProvider) Collect(ctx context.Context, opts happydns.CheckerOptions) (any, error) {
	// Read user-configurable options using the SDK helpers.
	// These helpers handle type coercion gracefully - the value may come as
	// a native Go type (in-process plugin) or as a JSON-decoded float64/string
	// (external HTTP mode). The helpers normalise both cases.
	message := "Hello from the dummy checker!"
	if v, ok := sdk.GetOption[string](opts, "message"); ok && v != "" {
		message = v
	}

	// Generate a random score between 0 and 100 to simulate a measurement.
	// In your real checker, replace this with actual monitoring logic.
	score := rand.Float64() * 100

	return &DummyData{
		Message:     message,
		Score:       score,
		CollectedAt: time.Now(),
	}, nil
}

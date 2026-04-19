package checker

import (
	"encoding/json"
	"time"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Provider returns a new dummy observation provider.
//
// The provider is the central object of a checker. It implements the
// ObservationProvider interface (required) and can optionally implement
// additional interfaces to unlock more features:
//
//   - CheckerDefinitionProvider  → exposes /definition and /evaluate endpoints
//   - CheckerMetricsReporter     → exposes /report (JSON metrics) endpoint
//   - CheckerHTMLReporter        → exposes /report (HTML) endpoint
//
// In this example, the provider implements all three optional interfaces
// so you can see how each one works.
func Provider() sdk.ObservationProvider {
	return &dummyProvider{}
}

// dummyProvider is the concrete type that satisfies the ObservationProvider
// interface and the optional reporter interfaces.
type dummyProvider struct{}

// Key returns the observation key for this provider. This must match the key
// used in your CheckerDefinition's ObservationKeys list so happyDomain knows
// which provider produces which data.
func (p *dummyProvider) Key() sdk.ObservationKey {
	return ObservationKeyDummy
}

// Definition implements sdk.CheckerDefinitionProvider.
// Returning a definition enables the /definition and /evaluate HTTP endpoints
// in the SDK server, and lets happyDomain discover this checker's metadata.
func (p *dummyProvider) Definition() *sdk.CheckerDefinition {
	return Definition()
}

// ExtractMetrics implements sdk.CheckerMetricsReporter.
// This is called when happyDomain (or the /report endpoint) needs to turn
// raw observation data into time-series metrics for graphing.
func (p *dummyProvider) ExtractMetrics(ctx sdk.ReportContext, collectedAt time.Time) ([]sdk.CheckMetric, error) {
	var data DummyData
	if err := json.Unmarshal(ctx.Data(), &data); err != nil {
		return nil, err
	}

	return []sdk.CheckMetric{
		{
			Name:      "dummy_score",
			Value:     data.Score,
			Unit:      "points",
			Timestamp: collectedAt,
		},
	}, nil
}

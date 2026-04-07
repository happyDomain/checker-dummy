package checker

import (
	"encoding/json"
	"time"

	"git.happydns.org/happyDomain/model"
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
func Provider() happydns.ObservationProvider {
	return &dummyProvider{}
}

// dummyProvider is the concrete type that satisfies the ObservationProvider
// interface and the optional reporter interfaces.
type dummyProvider struct{}

// Key returns the observation key for this provider. This must match the key
// used in your CheckerDefinition's ObservationKeys list so happyDomain knows
// which provider produces which data.
func (p *dummyProvider) Key() happydns.ObservationKey {
	return ObservationKeyDummy
}

// Definition implements happydns.CheckerDefinitionProvider.
// Returning a definition enables the /definition and /evaluate HTTP endpoints
// in the SDK server, and lets happyDomain discover this checker's metadata.
func (p *dummyProvider) Definition() *happydns.CheckerDefinition {
	return Definition()
}

// ExtractMetrics implements happydns.CheckerMetricsReporter.
// This is called when happyDomain (or the /report endpoint) needs to turn
// raw observation data into time-series metrics for graphing.
func (p *dummyProvider) ExtractMetrics(raw json.RawMessage, collectedAt time.Time) ([]happydns.CheckMetric, error) {
	var data DummyData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return []happydns.CheckMetric{
		{
			Name:      "dummy_score",
			Value:     data.Score,
			Unit:      "points",
			Timestamp: collectedAt,
		},
	}, nil
}

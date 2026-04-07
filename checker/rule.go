package checker

import (
	"context"
	"fmt"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Rule returns a new dummy check rule.
//
// A rule evaluates collected observation data and returns a status (OK,
// Warning, Critical, Error). Each checker can define multiple rules that
// inspect the same data from different angles.
func Rule() sdk.CheckRule {
	return &dummyRule{}
}

// dummyRule implements the sdk.CheckRule interface.
type dummyRule struct{}

// Name returns a unique, stable identifier for this rule. It is used as the
// "code" field in check results and stored in the database.
func (r *dummyRule) Name() string { return "dummy_score_check" }

// Description returns a human-readable summary of what this rule checks.
func (r *dummyRule) Description() string {
	return "Checks whether the dummy score is above the configured thresholds"
}

// ValidateOptions is called before evaluation to verify that the options are
// well-formed. Return an error to reject invalid configuration early, before
// any data collection happens.
func (r *dummyRule) ValidateOptions(opts sdk.CheckerOptions) error {
	warning := sdk.GetFloatOption(opts, "warningThreshold", 50)
	critical := sdk.GetFloatOption(opts, "criticalThreshold", 20)

	if warning < 0 || warning > 100 {
		return fmt.Errorf("warningThreshold must be between 0 and 100")
	}
	if critical < 0 || critical > 100 {
		return fmt.Errorf("criticalThreshold must be between 0 and 100")
	}
	if critical >= warning {
		return fmt.Errorf("criticalThreshold (%v) must be less than warningThreshold (%v)", critical, warning)
	}

	return nil
}

// Evaluate inspects the collected observation data and returns a CheckState.
//
// Parameters:
//   - ctx: context for cancellation.
//   - obs: an ObservationGetter to retrieve collected data by key.
//   - opts: the merged checker options.
//
// The ObservationGetter.Get method deserialises the stored JSON into your data
// struct. Always check the error: the observation may not be available if
// collection failed.
func (r *dummyRule) Evaluate(ctx context.Context, obs sdk.ObservationGetter, opts sdk.CheckerOptions) sdk.CheckState {
	// Retrieve the observation data by key.
	var data DummyData
	if err := obs.Get(ctx, ObservationKeyDummy, &data); err != nil {
		return sdk.CheckState{
			Status:  sdk.StatusError,
			Message: fmt.Sprintf("Failed to get dummy data: %v", err),
			Code:    "dummy_error",
		}
	}

	// Read thresholds from options.
	warningThreshold := sdk.GetFloatOption(opts, "warningThreshold", 50)
	criticalThreshold := sdk.GetFloatOption(opts, "criticalThreshold", 20)

	// Determine the status based on the score and thresholds.
	var status sdk.Status
	switch {
	case data.Score < criticalThreshold:
		status = sdk.StatusCrit
	case data.Score < warningThreshold:
		status = sdk.StatusWarn
	default:
		status = sdk.StatusOK
	}

	return sdk.CheckState{
		Status:  status,
		Message: fmt.Sprintf("Score: %.1f - %s", data.Score, data.Message),
		Code:    "dummy_score_check",
		Meta: map[string]any{
			"score":   data.Score,
			"message": data.Message,
		},
	}
}

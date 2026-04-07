package checker

import (
	"time"

	sdk "git.happydns.org/checker-sdk-go/checker"
)

// Version is the checker version reported in CheckerDefinition.Version.
//
// It defaults to "built-in", which is appropriate when the checker package is
// imported directly (built-in or plugin mode). Standalone binaries (like
// main.go) should override this from their own Version variable at the start
// of main(), which makes it easy for CI to inject a version with a single
// -ldflags "-X main.Version=..." flag instead of targeting the nested
// package path.
var Version = "built-in"

// Definition returns the CheckerDefinition for the dummy checker.
//
// A CheckerDefinition tells happyDomain everything it needs to know about
// your checker: its identity, where it can be applied, what options it
// accepts, what rules it provides, and how often it should run.
func Definition() *sdk.CheckerDefinition {
	return &sdk.CheckerDefinition{
		// ID is a unique, stable identifier for this checker. It is stored in
		// the database, so never change it after release.
		ID: "dummy",

		// Name is the human-readable label shown in the happyDomain UI.
		Name: "Dummy (example)",

		// Version is an optional version string for this checker. It is
		// surfaced in the UI/API and is useful to track which iteration of
		// your checker produced a given observation. The value is injected
		// at build time via -ldflags "-X .../checker.Version=...".
		Version: Version,

		// Availability controls where this checker appears in the UI.
		// A checker can apply at the domain level, zone level, or service
		// level. You can also restrict it to specific service types.
		//
		// Here we apply it at the domain level, which means users will see
		// this checker in the "Domain checks" section and it does not require
		// a specific service to be present.
		Availability: sdk.CheckerAvailability{
			ApplyToDomain: true,
		},

		// ObservationKeys lists the keys this checker produces. This ties
		// the definition to the provider(s) that generate the data.
		ObservationKeys: []sdk.ObservationKey{ObservationKeyDummy},

		// Options documents what configuration the checker accepts. Options
		// are grouped by audience (admin, user, domain, service, run):
		//
		//   - AdminOpts:   set once by the happyDomain administrator
		//   - UserOpts:    editable by end-users in the checker settings UI
		//   - DomainOpts:  auto-filled per domain (domain_name, etc.)
		//   - ServiceOpts: auto-filled per service (the service payload)
		//   - RunOpts:     set at collect-time only (e.g., overrides)
		//
		// Each option has an Id (used as the key in CheckerOptions), a Type
		// for the UI widget, a Label, and optionally a Default value.
		Options: sdk.CheckerOptionsDocumentation{
			UserOpts: []sdk.CheckerOptionDocumentation{
				{
					Id:          "message",
					Type:        "string",
					Label:       "Custom message",
					Description: "A message that will be included in the observation data.",
					Default:     "Hello from the dummy checker!",
				},
				{
					Id:          "warningThreshold",
					Type:        "number",
					Label:       "Warning threshold (score)",
					Description: "If the score drops below this value, the check status becomes Warning.",
					Default:     float64(50),
				},
				{
					Id:          "criticalThreshold",
					Type:        "number",
					Label:       "Critical threshold (score)",
					Description: "If the score drops below this value, the check status becomes Critical.",
					Default:     float64(20),
				},
			},
			DomainOpts: []sdk.CheckerOptionDocumentation{
				{
					Id:       "domain_name",
					Label:    "Domain name",
					AutoFill: sdk.AutoFillDomainName,
				},
			},
		},

		// Rules lists the evaluation rules provided by this checker. Each
		// rule will appear in the UI, and users can enable/disable them
		// individually.
		Rules: []sdk.CheckRule{
			Rule(),
		},

		// Interval specifies how often the check should run.
		Interval: &sdk.CheckIntervalSpec{
			Min:     1 * time.Minute,
			Max:     1 * time.Hour,
			Default: 5 * time.Minute,
		},

		// HasMetrics indicates that this checker can produce time-series
		// metrics (because our provider implements CheckerMetricsReporter).
		HasMetrics: true,
	}
}

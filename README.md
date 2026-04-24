# checker-dummy - How to Build a happyDomain Checker

This repository is a **fully working, educational example** of a happyDomain checker. It is intentionally simple: instead of performing real monitoring, it returns a random score and a user-configurable message. This lets you focus on learning the structure without dealing with external dependencies.

Use this as a template when you create your own checker.

---

## Table of Contents

1. [What is a Checker?](#what-is-a-checker)
2. [Architecture Overview](#architecture-overview)
3. [Repository Structure](#repository-structure)
4. [Step-by-Step Walkthrough](#step-by-step-walkthrough)
   - [Step 1: Define Your Data Types](#step-1-define-your-data-types)
   - [Step 2: Create the Provider](#step-2-create-the-provider)
   - [Step 3: Implement Data Collection](#step-3-implement-data-collection)
   - [Step 4: Describe Your Checker (Definition)](#step-4-describe-your-checker-definition)
   - [Step 5: Write Evaluation Rules](#step-5-write-evaluation-rules)
   - [Step 6: Wire It Up (main.go)](#step-6-wire-it-up-maingo)
   - [Step 7: Create the Plugin Entrypoint](#step-7-create-the-plugin-entrypoint)
5. [Optional: Standalone Human UI (`CheckerInteractive`)](#optional-standalone-human-ui-checkerinteractive)
6. [Running the Checker](#running-the-checker)
7. [Testing with curl](#testing-with-curl)
8. [Deploying to happyDomain](#deploying-to-happydomain)
9. [License & happyDomain compatibility](#license--happydomain-compatibility)
10. [Going Further](#going-further)

---

## What is a Checker?

A **checker** is a small, self-contained program that monitors one aspect of a domain's DNS infrastructure. happyDomain runs checkers periodically and displays their results in its dashboard.

Every checker does three things:

1. **Collect:** Gather raw observation data (e.g., ping a server, query an API, measure DNS response time).
2. **Evaluate:** Compare the collected data against user-defined thresholds to produce a status: OK, Warning, or Critical.
3. **Report** *(optional)*: Extract time-series metrics or generate HTML reports for the dashboard.


## Architecture Overview

A checker can run in three modes:

### Standalone HTTP Server (External Checker)

The checker runs as its own process and exposes an HTTP API. happyDomain communicates with it over the network. This is the most flexible option: you can write your checker in any language, deploy it independently, and scale it separately.

```
┌─────────────┐    HTTP     ┌─────────────────┐
│ happyDomain │ ──────────► │ checker-dummy   │
│   server    │ ◄────────── │ (this program)  │
└─────────────┘             └─────────────────┘
```

### In-Process Plugin

The checker is compiled as a Go plugin (`.so` file) and loaded directly into the happyDomain process. This is simpler to deploy (single binary) but requires the checker to be written in Go.

```
┌──────────────────────────────────────┐
│ happyDomain server                   │
│                                      │
│   ┌──────────────────────────────┐   │
│   │ checker-dummy.so (plugin)    │   │
│   │ checker-ping.so (plugin)     │   │
│   │ checker-matrix.so (plugin)   │   │
│   │ checker-....so (plugin)      │   │
│   └──────────────────────────────┘   │
└──────────────────────────────────────┘
```

### Built-in Checker

The checker package can be imported directly into the happyDomain server and registered at init time: no plugin loading, no separate process. This avoids the operational burden of Go's plugin system (matching toolchain versions, CGO, `.so` distribution) entirely.

```go
// happydomain/checkers/ping.go
package checkers

import (
    ping "git.happydns.org/checker-ping/checker"
    "git.happydns.org/internal/checker"
)

func init() {
    checker.RegisterObservationProvider(ping.Provider())
    checker.RegisterExternalizableChecker(ping.Definition())
}
```

This mode is reserved for checkers maintained as part of the happyDomain project itself. Or if you compile yourself your own version of happyDomain.

**Both standalone, plugin and built-in modes use the same checker code; only the entry point differs.**


## Repository Structure

```
checker-dummy/
├── main.go                 # Entry point for standalone HTTP server mode
├── checker/
│   ├── types.go            # Data structures (what the checker observes)
│   ├── provider.go         # The provider: glues everything together
│   ├── collect.go          # Collection logic (the actual monitoring)
│   ├── definition.go       # Checker metadata (options, rules, intervals)
│   └── rule.go             # Evaluation rules (OK / Warning / Critical)
├── plugin/
│   └── plugin.go           # Entry point for plugin mode
├── go.mod                  # Go module definition
├── Makefile                # Build targets
├── Dockerfile              # Container image
└── .gitignore
```

Each file has a single, clear responsibility. This is the recommended layout for all happyDomain checkers.

---


## Step-by-Step Walkthrough

### Step 1: Define Your Data Types

**File: `checker/types.go`**

Start by defining the data structure that your checker will produce during collection. This struct is serialised to JSON by the SDK, stored by happyDomain, and later deserialised during evaluation.

```go
const ObservationKeyDummy = "dummy"

type DummyData struct {
    Message     string    `json:"message"`
    Score       float64   `json:"score"`
    CollectedAt time.Time `json:"collected_at"`
}
```

Key points:
- **`ObservationKeyDummy`** is a unique string that identifies observations produced by this checker. Every checker needs at least one key.
- **Design for evaluation**: include everything your rules will need to decide OK/Warning/Critical. The evaluation step only sees this struct; it cannot re-collect data.

### Step 2: Create the Provider

**File: `checker/provider.go`**

The **provider** is the central object of your checker. It must implement the `ObservationProvider` interface:

```go
type ObservationProvider interface {
    Key() ObservationKey
    Collect(ctx context.Context, opts CheckerOptions) (any, error)
}
```

You can also implement optional interfaces to unlock additional features:

| Interface                     | What it enables                          |
|-------------------------------|------------------------------------------|
| `CheckerDefinitionProvider`   | `/definition` and `/evaluate` endpoints  |
| `CheckerMetricsReporter`      | `/report` endpoint (JSON metrics)        |
| `CheckerHTMLReporter`         | `/report` endpoint (HTML)                |
| `CheckerInteractive`          | `GET`/`POST /check` human-friendly HTML UI |

In this example, we implement all three optional interfaces:

```go
type dummyProvider struct{}

func (p *dummyProvider) Key() ObservationKey { return ObservationKeyDummy }
func (p *dummyProvider) Definition() *CheckerDefinition { return Definition() }
func (p *dummyProvider) ExtractMetrics(raw json.RawMessage, collectedAt time.Time) ([]CheckMetric, error) { ... }
```

The `Key()` method must return the same string as your `ObservationKeyDummy` constant.

### Step 3: Implement Data Collection

**File: `checker/collect.go`**

This is where the real work happens. The `Collect` method is called every time happyDomain runs your check.

```go
func (p *dummyProvider) Collect(ctx context.Context, opts CheckerOptions) (any, error) {
    // Read options using SDK helpers
    message := "Hello from the dummy checker!"
    if v, ok := sdk.GetOption[string](opts, "message"); ok && v != "" {
        message = v
    }

    // Do your monitoring work here!
    // In a real checker, you would: ping a server, query an API,
    // measure DNS response time, check TLS certificates, etc.
    score := rand.Float64() * 100

    return &DummyData{
        Message:     message,
        Score:       score,
        CollectedAt: time.Now(),
    }, nil
}
```

Key points:
- **Always honour `ctx`**: happyDomain may cancel long-running checks.
- **Use SDK option helpers** (`sdk.GetOption`, `sdk.GetFloatOption`, `sdk.GetIntOption`, `sdk.GetBoolOption`) to read options. They handle type coercion between in-process (native Go types) and HTTP mode (JSON-decoded types).
- **Return your data struct**: the SDK serialises it to JSON automatically.
- **Return an error** only if collection failed entirely. Partial results are fine.

### Step 4: Describe Your Checker (Definition)

**File: `checker/definition.go`**

The `CheckerDefinition` tells happyDomain everything about your checker:

```go
func Definition() *CheckerDefinition {
    return &CheckerDefinition{
        ID:      "dummy",           // Unique, stable identifier (never change after release)
        Name:    "Dummy (example)", // Human-readable label for the UI
        Version: Version,           // Optional; injected at build time via -ldflags

        Availability: CheckerAvailability{
            ApplyToDomain: true, // Show in the "Domain checks" section
        },

        ObservationKeys: []ObservationKey{ObservationKeyDummy},

        Options: CheckerOptionsDocumentation{
            UserOpts: []CheckerOptionDocumentation{
                {Id: "message", Type: "string", Label: "Custom message", Default: "Hello!"},
                {Id: "warningThreshold", Type: "number", Label: "Warning threshold", Default: float64(50)},
                ...
            },
        },

        Rules: []CheckRule{Rule()},

        Interval: &CheckIntervalSpec{
            Min: 1 * time.Minute, Max: 1 * time.Hour, Default: 5 * time.Minute,
        },

        HasMetrics: true,
    }
}
```

**Version**: declare a package-level `var Version = "built-in"` in your checker package and reference it from the definition. The default is fine when the package is imported directly (built-in or plugin mode). For the standalone binary, declare a separate `var Version = "custom-build"` in `main.go` and propagate it in `init()`:

```go
// main.go
var Version = "custom-build"

func init() {
    dummy.Version = Version
}
```

The CI can then override the standalone binary's version with a simple flag, without having to know the nested package path:

```bash
go build -ldflags "-X main.Version=$(git describe --tags)" .
```

**Availability**: choose where your checker appears:

| Field            | When to use                                         |
|------------------|-----------------------------------------------------|
| `ApplyToDomain`  | The check applies to the entire domain              |
| `ApplyToZone`    | The check applies to a specific DNS zone            |
| `ApplyToService` | The check applies to a specific service (e.g., A/AAAA records). Use `LimitToServices` to restrict which service types. |

**Options**: grouped by audience:

| Group         | Who sets it          | Example                              |
|---------------|----------------------|--------------------------------------|
| `AdminOpts`   | happyDomain admin    | API endpoint URL                     |
| `UserOpts`    | End-user in the UI   | Thresholds, count, custom messages   |
| `DomainOpts`  | Auto-filled per domain | `domain_name` (via `AutoFill`)     |
| `ServiceOpts` | Auto-filled per service | The service payload (via `AutoFill`) |
| `RunOpts`     | Set at collect-time  | Runtime overrides                    |

**Option types** for the UI widget: `"string"`, `"number"`, `"uint"`, `"bool"`. You can also provide `Choices` for dropdown menus.

### Step 5: Write Evaluation Rules

**File: `checker/rule.go`**

A rule implements the `CheckRule` interface:

```go
type CheckRule interface {
    Name() string
    Description() string
    Evaluate(ctx context.Context, obs ObservationGetter, opts CheckerOptions) []CheckState
}
```

Optionally, your rule can also implement `ValidateOptions(opts) error` for early validation.

The `Evaluate` method receives an `ObservationGetter` to retrieve the collected data and returns a **slice** of `CheckState`, one entry per element being evaluated:

```go
func (r *dummyRule) Evaluate(ctx context.Context, obs ObservationGetter, opts CheckerOptions) []CheckState {
    var data DummyData
    if err := obs.Get(ctx, ObservationKeyDummy, &data); err != nil {
        return []CheckState{{Status: StatusError, Message: "..."}}
    }

    warningThreshold := sdk.GetFloatOption(opts, "warningThreshold", 50)
    criticalThreshold := sdk.GetFloatOption(opts, "criticalThreshold", 20)

    switch {
    case data.Score < criticalThreshold:
        return []CheckState{{Status: StatusCrit, ...}}
    case data.Score < warningThreshold:
        return []CheckState{{Status: StatusWarn, ...}}
    default:
        return []CheckState{{Status: StatusOK, ...}}
    }
}
```

**Contract**: `Evaluate` must return at least one state. If a rule has nothing to evaluate, return a single `CheckState` describing that fact (typically `StatusInfo` or `StatusOK`). The SDK server injects a `StatusUnknown` placeholder if a rule returns an empty or nil slice.

**The `CheckState` struct**:

```go
type CheckState struct {
    Status   Status
    Message  string
    RuleName string         // set automatically by the server, do not set yourself
    Code     string         // optional, use to distinguish kinds of finding within one rule
    Subject  string         // opaque per-element identifier (hostname, cert serial, …)
    Meta     map[string]any
}
```

- **`Subject`** identifies the element a state refers to (a hostname, a certificate serial, a nameserver FQDN, …). Leave empty for rules that produce a single global result. Do **not** repeat the subject inside `Message`, the UI renders it separately.
- **`RuleName`** is stamped automatically by the server with `rule.Name()` on every returned state. UIs should use `RuleName` (not `Code`) to group, filter, or offer "disable this rule" controls.
- **`Code`** is left untouched by the server. Set it only when your rule emits several kinds of finding (e.g. `too_many_lookups` vs `syntax_error`).

**One state per subject**: a rule that iterates over N elements should emit N states (one per `Subject`) instead of concatenating them into a single `Message`:

```go
func (r *CertExpiryRule) Evaluate(...) []CheckState {
    out := make([]CheckState, 0, len(certs))
    for _, cert := range certs {
        s := evalCert(cert)
        s.Subject = cert.Host
        out = append(out, s)
    }
    if len(out) == 0 {
        return []CheckState{{Status: StatusInfo, Message: "no certificate to evaluate"}}
    }
    return out
}
```

**Status values**: `StatusOK`, `StatusWarn`, `StatusCrit`, `StatusError`, `StatusUnknown`, `StatusInfo`.

You can define **multiple rules** per checker. Each rule evaluates the same collected data from a different angle. Users can enable/disable rules individually in the UI.

### Step 6: Wire It Up (main.go)

**File: `main.go`**

The standalone entry point is minimal; the SDK does all the heavy lifting:

```go
func main() {
    flag.Parse()

    // Propagate the plugin's version to the checker package.
    dummy.Version = Version

    server := sdk.NewServer(dummy.Provider())
    server.ListenAndServe(*listenAddr)
}
```

`sdk.NewServer` inspects your provider and automatically registers HTTP endpoints based on which interfaces it implements:

| Endpoint           | Always | Requires                     |
|--------------------|--------|------------------------------|
| `GET /health`      | Yes    | -                            |
| `POST /collect`    | Yes    | -                            |
| `GET /definition`  | -      | `CheckerDefinitionProvider`  |
| `POST /evaluate`   | -      | `CheckerDefinitionProvider`  |
| `POST /report`     | -      | `CheckerMetricsReporter` or `CheckerHTMLReporter` |
| `GET`/`POST /check` | -     | `CheckerInteractive`         |

### Step 7: Create the Plugin Entrypoint

**File: `plugin/plugin.go`**

For in-process plugin mode, the entrypoint must be a `package main` that exposes a `NewCheckerPlugin` symbol. happyDomain opens the `.so` file with `plugin.Open`, looks up that symbol, and calls it to obtain the checker definition and its observation provider, which the host then registers in its own global registries.

```go
package main

import (
    sdk "git.happydns.org/checker-sdk-go/checker"
    dummy "git.happydns.org/checker-dummy/checker"
)

var Version = "custom-build"

func NewCheckerPlugin() (*sdk.CheckerDefinition, sdk.ObservationProvider, error) {
    // Propagate the plugin's version to the checker package.
    dummy.Version = Version
    return dummy.Definition(), dummy.Provider(), nil
}
```

Build the plugin with:

```bash
go build -buildmode=plugin -o checker-dummy.so ./plugin
```

Then drop the resulting `checker-dummy.so` into one of happyDomain's configured plugin directories. It will be picked up at startup.

---


## Optional: Standalone Human UI (`CheckerInteractive`)

The SDK provides an optional `CheckerInteractive` interface that exposes a browser-friendly `/check` route, letting your checker be used as a standalone DNS-probing tool without a happyDomain instance in front of it.

```go
type CheckerInteractive interface {
    RenderForm() []CheckerOptionField
    ParseForm(r *http.Request) (CheckerOptions, error)
}
```

When a provider implements it, `NewServer` automatically registers:

- `GET /check`, renders an HTML form derived from `RenderForm()`.
- `POST /check`, calls `ParseForm`, runs the standard `Collect` → `Evaluate` → `GetHTMLReport` / `ExtractMetrics` pipeline, and returns a consolidated HTML page (states table, metrics table, sandboxed iframe around the HTML report).

### Why it exists

Over the HTTP `/evaluate` endpoint, happyDomain fills `AutoFill*`-backed options (zone records, service payload, …) from its execution context. A human hitting `/check` has no such host, `ParseForm` is where the checker does whatever lookups are needed (typically direct DNS queries) to turn a minimal human input (e.g. a domain name) into the full `CheckerOptions` that `Collect` expects.

### When to implement it

- You want the checker to be usable as a standalone DNS-probing tool (debug, demo, one-off runs) without a happyDomain instance.
- You are fine doing the auto-fill work yourself from the user's inputs. Checkers whose `Collect` intrinsically requires data only happyDomain can provide (e.g. a full zone diff) should skip this.

### Minimal implementation

```go
func (p *dummyProvider) RenderForm() []sdk.CheckerOptionField {
    return []sdk.CheckerOptionField{
        {Id: "message", Type: "string", Label: "Custom message",
         Placeholder: "Hello!", Required: false},
    }
}

func (p *dummyProvider) ParseForm(r *http.Request) (sdk.CheckerOptions, error) {
    return sdk.CheckerOptions{
        "message": strings.TrimSpace(r.FormValue("message")),
    }, nil
}
```

Returning an error from `ParseForm` re-renders the form with the error message displayed so the user can correct and resubmit.

---


## Running the Checker

### Build and run locally

```bash
make build
./checker-dummy -listen :8080
```

### Docker

```bash
make docker
docker run -p 8080:8080 happydomain/checker-dummy
```

---


## Testing with curl

### Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Get the checker definition

```bash
curl http://localhost:8080/definition
```

### Collect an observation

```bash
curl -X POST http://localhost:8080/collect \
  -H "Content-Type: application/json" \
  -d '{
    "key": "dummy",
    "options": {
      "message": "Testing my checker!"
    }
  }'
```

Response:
```json
{
  "data": {
    "message": "Testing my checker!",
    "score": 73.2,
    "collected_at": "2026-01-15T10:30:00Z"
  }
}
```

### Evaluate observations

```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "observations": {
      "dummy": "{\"message\":\"test\",\"score\":42.5,\"collected_at\":\"2026-01-15T10:30:00Z\"}"
    },
    "options": {
      "warningThreshold": 50,
      "criticalThreshold": 20
    }
  }'
```

Response (score 42.5 is below the warning threshold of 50):
```json
{
  "states": [
    {
      "status": 3,
      "message": "Score: 42.5 - test",
      "rule_name": "dummy_score_check",
      "code": "dummy_score_check"
    }
  ]
}
```

Each entry in `states` carries a `rule_name` (server-stamped) and may include a `subject` field when the rule evaluates multiple elements.

Status codes: `1` = OK, `3` = Warning, `4` = Critical.

### Extract metrics

```bash
curl -X POST http://localhost:8080/report \
  -H "Content-Type: application/json" \
  -d '{
    "data": "{\"message\":\"test\",\"score\":73.2,\"collected_at\":\"2026-01-15T10:30:00Z\"}"
  }'
```

---


## Deploying to happyDomain

### As an external checker (recommended)

1. Deploy your checker as a standalone service (Docker, systemd, etc.).
2. In happyDomain, set the checker's `endpoint` admin option to its URL (e.g., `http://checker-dummy:8080`).
3. happyDomain will call `/collect`, `/evaluate`, and `/report` automatically.

### As an in-process plugin

1. Build the plugin:
   ```bash
   go build -buildmode=plugin -o checker-dummy.so ./plugin
   ```
2. Copy `checker-dummy.so` into one of the directories listed in happyDomain's `PluginsDirectories` configuration.
3. Restart happyDomain. At startup it scans those directories, opens each `.so`, looks up the `NewCheckerPlugin` symbol, and registers the returned definition and provider.

---

## License & happyDomain compatibility

This template is released under the **MIT License**, so you're free to use it as a starting point for any checker, including proprietary ones.

The types and helpers your checker depends on live in [`checker-sdk-go`](https://git.happydns.org/checker-sdk-go), a separate module released under the **Apache License 2.0**. happyDomain itself depends on the same SDK, so plugins and the host share a common, permissively licensed contract instead of linking against AGPL code.

**What this means for the deployment mode you choose:**

- **Standalone HTTP checker:** your checker is a separate process communicating with happyDomain over the network. It is *not* a derivative work of happyDomain and you can license it however you want (proprietary, MIT, GPL, anything).
- **In-process plugin (`.so`):** your checker is loaded into the happyDomain process via `plugin.Open`, but it only links against the Apache-licensed SDK, not against any AGPL code. You are free to license your plugin however you want.
- **Built-in checker** (imported directly into the happyDomain source tree): same as above on the linking side. Built-in checkers maintained inside the happyDomain repository are conventionally distributed under AGPL-3.0 to stay consistent with the rest of the project, but this is a project policy, not a legal requirement coming from the SDK.

If your checker imports anything *else* from the happyDomain repository (for example service abstractions like `happydns.ServiceMessage`), then that code *is* AGPL-licensed and the AGPL constraint comes back. The SDK alone is safe; the rest of happyDomain is not.

---


## Going Further

Now that you understand the structure, here are ideas for your own checker:

- **SMTP checker:** connect to a mail server and verify it responds correctly to EHLO.
- **DNS checker:** query specific DNS record types and verify the response matches expectations.
- **HTTP checker:** send an HTTP request to a domain's web server and check the status code, response time, ...
- **Business logic:** probe your own application from the outside and verify it behaves as expected, e.g. log into your SaaS with a synthetic account and check that the dashboard loads, place a test order and confirm it reaches the order pipeline, hit an internal health endpoint that aggregates queue depth / worker lag / replication status, or check that a license server still hands out valid tokens. This turns happyDomain into a lightweight synthetic-monitoring dashboard for your own services.

For a real-world example, look at [checker-ping](https://git.happydns.org/checker-ping), which implements ICMP ping monitoring with multiple targets, packet loss detection, and RTT metrics.

### Tips

- Keep `Collect` focused on data gathering. Put all threshold logic in `Evaluate`.
- Design your data struct to hold everything rules need; evaluation cannot re-collect.
- Use `sdk.GetFloatOption` / `sdk.GetIntOption` / `sdk.GetBoolOption` instead of raw type assertions. They handle the JSON/native type mismatch transparently.
- Always honour the `context.Context`: set timeouts and check for cancellation.
- Return partial results from `Collect` when possible (only return an error if the entire collection failed).

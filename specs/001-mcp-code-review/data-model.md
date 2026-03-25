# Data Model: MCP Code Review Server

**Branch**: `001-mcp-code-review` | **Phase**: 1

---

## Core Types

### ReviewInput

Represents the normalized input to any review operation, after the MCP tool layer has
resolved the diff or code string.

```go
// internal/review/types.go

type ReviewInput struct {
    // Code is the diff text or source code to review.
    Code string

    // Language is an optional hint about the programming language or diff format.
    // Examples: "go", "python", "unified-diff"
    Language string

    // Context is optional human-readable context (e.g., commit message, file paths).
    Context string

    // Provider is the resolved provider name: "anthropic" | "openai" | "google".
    Provider string

    // Model is the resolved model identifier. Empty string means use provider default.
    Model string
}
```

**Validation rules**:
- `Code` MUST be non-empty; empty string returns an error before any provider call.
- `Provider` MUST be one of: `"anthropic"`, `"openai"`, `"google"` after resolution.
- `Model` may be empty (provider fills in its default).

---

### ReviewResult

The structured output returned from any review operation.

```go
// internal/review/types.go

type ReviewResult struct {
    // Summary is a concise overall assessment (1-3 sentences).
    Summary string `json:"summary"`

    // Findings is the list of specific observations from the review.
    Findings []Finding `json:"findings"`
}

type Finding struct {
    // Severity is one of: "issue" | "suggestion" | "praise"
    Severity string `json:"severity"`

    // File is the file path the finding relates to. May be empty for overall findings.
    File string `json:"file,omitempty"`

    // Lines is a human-readable line range string, e.g. "12-15" or "42". May be empty.
    Lines string `json:"lines,omitempty"`

    // Message is the human-readable review comment.
    Message string `json:"message"`
}
```

**Validation rules**:
- `Severity` MUST be one of `"issue"`, `"suggestion"`, `"praise"`. Invalid values from
  model output are normalized to `"suggestion"`.
- `Message` MUST be non-empty. Empty messages are filtered from findings.
- `Summary` MUST be non-empty; if model returns empty summary, a default is substituted.

---

### Provider Interface

```go
// internal/provider/provider.go

// Provider is the abstraction over AI backends.
type Provider interface {
    // Review submits the given input for review and returns structured findings.
    Review(ctx context.Context, input review.ReviewInput) (*review.ReviewResult, error)

    // Name returns the canonical provider identifier ("anthropic" | "openai" | "google").
    Name() string

    // DefaultModel returns the model used when ReviewInput.Model is empty.
    DefaultModel() string
}
```

**Implementations**:
- `AnthropicProvider` — wraps `anthropics/anthropic-sdk-go`
- `OpenAIProvider` — wraps `openai/openai-go/v3`
- `GoogleProvider` — wraps `google.golang.org/genai`

---

### Provider Registry

```go
// internal/provider/registry.go

// Registry maps provider name strings to constructor functions.
type Registry struct {
    constructors map[string]ConstructorFunc
}

// ConstructorFunc creates a Provider from an optional model override.
type ConstructorFunc func(model string) (Provider, error)
```

**Registration**: All three providers are registered at server startup. Constructor
functions read API keys from environment variables:
- Anthropic: `ANTHROPIC_API_KEY`
- OpenAI: `OPENAI_API_KEY`
- Google: `GOOGLE_API_KEY` (or `GOOGLE_APPLICATION_CREDENTIALS` for ADC)

**Resolution order** (when provider is unspecified): checks `ANTHROPIC_API_KEY` →
`OPENAI_API_KEY` → `GOOGLE_API_KEY`; first non-empty key wins.

---

### MCP Tool Input Structs

These structs define the schema for each registered MCP tool.

```go
// internal/server/tools.go

type ReviewStagedInput struct {
    RepoPath string `json:"repo_path,omitempty" description:"Absolute path to git repository root. Defaults to current working directory."`
    Provider string `json:"provider,omitempty"  description:"AI provider: anthropic | openai | google. Defaults to first available API key."`
    Model    string `json:"model,omitempty"     description:"Model identifier. Defaults to provider's recommended model."`
}

type ReviewUnstagedInput struct {
    RepoPath string `json:"repo_path,omitempty"`
    Provider string `json:"provider,omitempty"`
    Model    string `json:"model,omitempty"`
}

type ReviewCommitInput struct {
    SHA      string `json:"sha"                 description:"Git commit SHA (full or abbreviated ≥7 chars). Required."`
    RepoPath string `json:"repo_path,omitempty"`
    Provider string `json:"provider,omitempty"`
    Model    string `json:"model,omitempty"`
}

type ReviewCodeInput struct {
    Code     string `json:"code"                description:"Source code or text to review. Required."`
    Language string `json:"language,omitempty"  description:"Optional language hint, e.g. 'go', 'python'."`
    Provider string `json:"provider,omitempty"`
    Model    string `json:"model,omitempty"`
}
```

---

### Git Diff Result

Internal type returned by git extraction functions.

```go
// internal/git/git.go

type DiffResult struct {
    // Diff is the raw unified diff text.
    Diff string

    // Files is the list of files touched (parsed from diff headers).
    Files []string

    // IsBinary reports whether any binary files were encountered (and skipped).
    IsBinary bool
}
```

**State transitions**:
- Empty `Diff` + no error → caller returns "nothing to review" message to MCP client.
- `IsBinary == true` → a note is prepended to the review summary.

---

## Entity Relationships

```
MCP Tool Call
    │
    ▼
ReviewXxxInput  ──(validate)──►  ReviewInput
                                      │
                              ┌───────┴────────┐
                              │                │
                         git.DiffResult    (for ReviewCode:
                         (for git tools)    Code string directly)
                              │
                              └───────┬────────┘
                                      │
                                 provider.Provider
                                      │
                                      ▼
                                 ReviewResult
                                      │
                                      ▼
                              MCP CallToolResult (JSON text)
```

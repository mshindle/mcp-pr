package review

// ReviewInput is the normalized input to any review operation.
type ReviewInput struct {
	// Code is the diff text or source code to review. Must be non-empty.
	Code string

	// Language is an optional hint about the programming language or diff format.
	Language string

	// Context is optional human-readable context (e.g., commit message, file paths).
	Context string

	// Provider is the resolved provider name: "anthropic" | "openai" | "google".
	Provider string

	// Model is the resolved model identifier. Empty string means use provider default.
	Model string
}

// ReviewResult is the structured output returned from any review operation.
type ReviewResult struct {
	Summary  string    `json:"summary"`
	Findings []Finding `json:"findings"`
}

// Finding is a single observation from the review.
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

// validSeverities lists the accepted severity values.
var validSeverities = map[string]bool{
	"issue":      true,
	"suggestion": true,
	"praise":     true,
}

// NormalizeSeverity returns the severity unchanged if valid, or "suggestion" if not.
func NormalizeSeverity(s string) string {
	if validSeverities[s] {
		return s
	}
	return "suggestion"
}

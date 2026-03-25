package provider

import (
	"encoding/json"
	"testing"

	"github.com/mshindle/mcp-pr/internal/review"
)

func TestBuildSystemPrompt_NotEmpty(t *testing.T) {
	p := review.BuildSystemPrompt()
	if p == "" {
		t.Error("expected non-empty system prompt")
	}
}

func TestBuildUserMessage_IncludesCode(t *testing.T) {
	input := review.ReviewInput{
		Code:     "func main() {}",
		Language: "go",
	}
	msg := review.BuildUserMessage(input)
	if msg == "" {
		t.Error("expected non-empty user message")
	}
	// Must include the code
	if !containsStr(msg, "func main()") {
		t.Errorf("user message does not contain code: %s", msg)
	}
}

func TestBuildUserMessage_WithContext(t *testing.T) {
	input := review.ReviewInput{
		Code:    "x = 1",
		Context: "fix bug in loop",
	}
	msg := review.BuildUserMessage(input)
	if !containsStr(msg, "fix bug in loop") {
		t.Errorf("user message does not contain context: %s", msg)
	}
}

func TestParseReviewResult_Valid(t *testing.T) {
	raw := `{"summary":"looks good","findings":[{"severity":"issue","file":"main.go","lines":"5","message":"unused var"}]}`
	result, err := review.ParseReviewResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "looks good" {
		t.Errorf("unexpected summary: %s", result.Summary)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	if result.Findings[0].Severity != "issue" {
		t.Errorf("unexpected severity: %s", result.Findings[0].Severity)
	}
}

func TestParseReviewResult_NormalizesInvalidSeverity(t *testing.T) {
	raw := `{"summary":"ok","findings":[{"severity":"critical","message":"bad"}]}`
	result, err := review.ParseReviewResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Findings[0].Severity != "suggestion" {
		t.Errorf("expected severity 'suggestion', got: %s", result.Findings[0].Severity)
	}
}

func TestParseReviewResult_FiltersEmptyMessages(t *testing.T) {
	raw := `{"summary":"ok","findings":[{"severity":"issue","message":""},{"severity":"praise","message":"nice"}]}`
	result, err := review.ParseReviewResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 finding after filtering empty messages, got %d", len(result.Findings))
	}
}

func TestParseReviewResult_DefaultSummaryWhenEmpty(t *testing.T) {
	raw := `{"summary":"","findings":[]}`
	result, err := review.ParseReviewResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary == "" {
		t.Error("expected non-empty default summary")
	}
}

func TestParseReviewResult_FallbackOnInvalidJSON(t *testing.T) {
	result, err := review.ParseReviewResult("not json at all")
	if err != nil {
		t.Fatalf("fallback should not return error: %v", err)
	}
	if result.Summary == "" {
		t.Error("expected non-empty summary in fallback result")
	}
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 fallback finding, got %d", len(result.Findings))
	}
}

func TestReviewResultJSON_RoundTrip(t *testing.T) {
	orig := &review.ReviewResult{
		Summary: "test",
		Findings: []review.Finding{
			{Severity: "praise", Message: "clean"},
		},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := review.ParseReviewResult(string(data))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Summary != orig.Summary {
		t.Errorf("summary mismatch: %s vs %s", got.Summary, orig.Summary)
	}
}

func containsStr(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && (len(s) >= len(sub)) && stringContains(s, sub)
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

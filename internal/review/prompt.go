package review

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BuildSystemPrompt returns the system instruction for the code review LLM.
func BuildSystemPrompt() string {
	return `You are an expert code reviewer. Your task is to review the provided code or diff and return a structured JSON response.

Return ONLY valid JSON — no markdown fences, no prose before or after. The JSON must conform to this schema:
{
  "summary": "<concise overall assessment, 1-3 sentences>",
  "findings": [
    {
      "severity": "<issue|suggestion|praise>",
      "file": "<file path, optional>",
      "lines": "<line range e.g. '12-15' or '42', optional>",
      "message": "<specific, actionable review comment>"
    }
  ]
}

Guidelines:
- Use "issue" for bugs, security problems, or correctness errors that should be fixed.
- Use "suggestion" for improvements to style, readability, or best practices.
- Use "praise" for well-written code worth highlighting.
- Be specific: reference file names and line numbers when available.
- If there is nothing to review, return a summary explaining why and an empty findings array.`
}

// BuildUserMessage constructs the user-turn message from a ReviewInput.
func BuildUserMessage(input ReviewInput) string {
	var sb strings.Builder

	if input.Context != "" {
		sb.WriteString("Context: ")
		sb.WriteString(input.Context)
		sb.WriteString("\n\n")
	}

	if input.Language != "" {
		sb.WriteString(fmt.Sprintf("Language/format: %s\n\n", input.Language))
	}

	sb.WriteString("Please review the following:\n\n")
	sb.WriteString(input.Code)

	return sb.String()
}

// ParseReviewResult parses a JSON string into a ReviewResult, applying validation and
// normalization rules. On JSON parse failure it returns a single-finding fallback result.
func ParseReviewResult(raw string) (*ReviewResult, error) {
	raw = strings.TrimSpace(raw)

	// Strip markdown code fences if the model included them.
	if strings.HasPrefix(raw, "```") {
		raw = stripFences(raw)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Fallback: surface the raw text as a single suggestion finding.
		return &ReviewResult{
			Summary: "Review returned non-JSON output; see findings for raw text.",
			Findings: []Finding{
				{
					Severity: "suggestion",
					Message:  raw,
				},
			},
		}, nil
	}

	// Normalize and validate.
	if result.Summary == "" {
		result.Summary = "Review complete. See findings for details."
	}

	var cleaned []Finding
	for _, f := range result.Findings {
		if f.Message == "" {
			continue
		}
		f.Severity = NormalizeSeverity(f.Severity)
		cleaned = append(cleaned, f)
	}
	result.Findings = cleaned

	return &result, nil
}

// stripFences removes leading/trailing markdown code fences from a string.
func stripFences(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "```") {
			continue
		}
		if i == len(lines)-1 && strings.TrimSpace(line) == "```" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

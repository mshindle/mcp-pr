package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/mshindle/mcp-pr/internal/review"
	"google.golang.org/genai"
)

const defaultGoogleModel = "gemini-2.0-flash"

// GoogleProvider implements Provider using the Google Generative AI API.
type GoogleProvider struct {
	client *genai.Client
	model  string
}

// NewGoogleProvider creates a GoogleProvider. Returns an error if GOOGLE_API_KEY is not set.
func NewGoogleProvider(modelOverride string) (Provider, error) {
	key := os.Getenv("GOOGLE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("provider 'google': missing API key; set GOOGLE_API_KEY")
	}

	model := modelOverride
	if model == "" {
		model = os.Getenv("MCP_REVIEW_GOOGLE_MODEL")
	}
	if model == "" {
		model = defaultGoogleModel
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("provider 'google': initialize client: %w", err)
	}

	return &GoogleProvider{client: client, model: model}, nil
}

func (p *GoogleProvider) Name() string         { return "google" }
func (p *GoogleProvider) DefaultModel() string { return p.model }

func (p *GoogleProvider) Review(ctx context.Context, input review.ReviewInput) (*review.ReviewResult, error) {
	model := input.Model
	if model == "" {
		model = p.model
	}

	systemPrompt := review.BuildSystemPrompt()
	userMessage := review.BuildUserMessage(input)

	resp, err := p.client.Models.GenerateContent(ctx, model,
		[]*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: userMessage}}},
		},
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("provider 'google' returned error: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("provider 'google' returned error: empty response")
	}

	raw := resp.Candidates[0].Content.Parts[0].Text
	return review.ParseReviewResult(raw)
}

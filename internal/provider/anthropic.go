package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/mshindle/mcp-pr/internal/review"
)

const defaultAnthropicModel = "claude-sonnet-4-6"

// AnthropicProvider implements Provider using the Anthropic Messages API.
type AnthropicProvider struct {
	client anthropic.Client
	model  string
}

// NewAnthropicProvider creates an AnthropicProvider. Returns an error if ANTHROPIC_API_KEY
// is not set (unless a key is embedded in the environment at SDK init time).
func NewAnthropicProvider(modelOverride string) (Provider, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("provider 'anthropic': missing API key; set ANTHROPIC_API_KEY")
	}

	model := modelOverride
	if model == "" {
		model = os.Getenv("MCP_REVIEW_ANTHROPIC_MODEL")
	}
	if model == "" {
		model = defaultAnthropicModel
	}

	client := anthropic.NewClient(option.WithAPIKey(key))
	return &AnthropicProvider{client: client, model: model}, nil
}

func (p *AnthropicProvider) Name() string         { return "anthropic" }
func (p *AnthropicProvider) DefaultModel() string { return p.model }

func (p *AnthropicProvider) Review(ctx context.Context, input review.ReviewInput) (*review.ReviewResult, error) {
	model := input.Model
	if model == "" {
		model = p.model
	}

	systemPrompt := review.BuildSystemPrompt()
	userMessage := review.BuildUserMessage(input)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("provider 'anthropic' returned error: %w", err)
	}

	// Extract text from first content block.
	var raw string
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			raw = tb.Text
			break
		}
	}

	return review.ParseReviewResult(raw)
}

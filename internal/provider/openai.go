package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/mshindle/mcp-pr/internal/review"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const defaultOpenAIModel = "gpt-4o"

// OpenAIProvider implements Provider using the OpenAI Chat Completions API.
type OpenAIProvider struct {
	client openai.Client
	model  string
}

// NewOpenAIProvider creates an OpenAIProvider. Returns an error if OPENAI_API_KEY is not set.
func NewOpenAIProvider(modelOverride string) (Provider, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("provider 'openai': missing API key; set OPENAI_API_KEY")
	}

	model := modelOverride
	if model == "" {
		model = os.Getenv("MCP_REVIEW_OPENAI_MODEL")
	}
	if model == "" {
		model = defaultOpenAIModel
	}

	client := openai.NewClient(option.WithAPIKey(key))
	return &OpenAIProvider{client: client, model: model}, nil
}

func (p *OpenAIProvider) Name() string         { return "openai" }
func (p *OpenAIProvider) DefaultModel() string { return p.model }

func (p *OpenAIProvider) Review(ctx context.Context, input review.ReviewInput) (*review.ReviewResult, error) {
	model := input.Model
	if model == "" {
		model = p.model
	}

	systemPrompt := review.BuildSystemPrompt()
	userMessage := review.BuildUserMessage(input)

	result, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userMessage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("provider 'openai' returned error: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("provider 'openai' returned error: empty response")
	}

	raw := result.Choices[0].Message.Content
	return review.ParseReviewResult(raw)
}

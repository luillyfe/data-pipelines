package llm

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/liushuangls/go-anthropic/v2"
)

type AnthropicLLM struct {
	ModelName string
	client    *anthropic.Client
}

func NewAnthropicLLM(modelName string) LanguageModel {
	if modelName == "" {
		modelName = anthropic.ModelClaudeInstant1Dot2
	}

	CLAUDE_API_KEY := os.Getenv("CLAUDE_API_KEY")
	return &AnthropicLLM{
		ModelName: modelName,
		client:    anthropic.NewClient(CLAUDE_API_KEY),
	}
}

func (a *AnthropicLLM) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Using chat completion
	resp, err := a.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: a.ModelName,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			return "", fmt.Errorf("anthropic API error, type: %s, message: %s", e.Type, e.Message)
		}
		return "", fmt.Errorf("anthropic API error: %w", err)
	}

	// return generated text
	return *resp.Content[0].Text, nil
}

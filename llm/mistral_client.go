package llm

import (
	"context"
	"fmt"

	"github.com/gage-technologies/mistral-go"
)

type mistralLLM struct {
	ModelName string
	client    *mistral.MistralClient
}

func NewMistralLLM(modelName string) LanguageModel {
	if modelName == "" {
		modelName = "mistral-small-latest"
	}

	return &mistralLLM{
		ModelName: modelName,
		client:    mistral.NewMistralClientDefault(""),
	}
}

func (m *mistralLLM) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Using chat completion
	resp, err := m.client.Chat(m.ModelName, []mistral.ChatMessage{{Content: prompt, Role: mistral.RoleUser}}, &mistral.ChatRequestParams{
		Temperature: 0.7,
		MaxTokens:   512,
		TopP:        1,
	})
	if err != nil {
		return "", fmt.Errorf("error getting chat completion: %w", err)
	}

	// return generated text
	return resp.Choices[0].Message.Content, nil
}

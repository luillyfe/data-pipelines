package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/gage-technologies/mistral-go"
)

// LLMClient is a PTransform that enriches questions with multiple choice options using an LLM
type LLMClient struct {
	ModelName string
	client    *mistral.MistralClient
	once      sync.Once
}

func (llm *LLMClient) ProcessElement(ctx context.Context, question *Question) (*MultipleChoiceQuestion, error) {
	// Lazy initialization
	llm.once.Do(func() {
		llm.setupClient()
	})

	prompt := fmt.Sprintf(`Analyze the following question and provide: 1. Four choices that could be used as answers. 2. Indicate which choice is correct. Question: %s Respond in the following format: Choices: A. [choice1] B. [choice2] C. [choice3] D. [choice4] Answer: [A/B/C/D]`, question.Text)

	// Using Chat completion
	chatRes, err := llm.client.Chat(llm.ModelName, []mistral.ChatMessage{{Content: prompt, Role: mistral.RoleUser}}, &mistral.ChatRequestParams{
		Temperature: 0.7,
		MaxTokens:   512,
		TopP:        1,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting chat completion: %w", err)
	}

	// Parse Question string to MultipleChoiceQuestion.
	mQuestion, err := parseLLMOutput(chatRes.Choices[0].Message.Content, question)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	return mQuestion, nil
}

func init() {
	// type arguments [context.Context, *Question, *MultipleChoiceQuestion, error]
	register.DoFn2x2(&LLMClient{})
}

func (llm *LLMClient) setupClient() {
	llm.once.Do(func() {
		llm.client = mistral.NewMistralClientDefault("")
	})
}

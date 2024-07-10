package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/liushuangls/go-anthropic/v2"
)

// LLMClient is a PTransform that enriches questions with multiple choice options using an LLM
type LLMClient struct {
	ModelName string
	client    *anthropic.Client
	once      sync.Once
}

func (llm *LLMClient) ProcessElement(ctx context.Context, question *Question) (*MultipleChoiceQuestion, error) {
	// Lazy initialization
	llm.once.Do(func() {
		llm.setupClient()
	})

	prompt := fmt.Sprintf(`Analyze the following question and provide: 1. Four choices that could be used as answers. 2. Indicate which choice is correct. 3. The choices makes the question easy to answer. Question: %s , the question belongs to the following exam's sections: %s and has ben tag with the following labels: %s. Please keep your output scoped to the Google Cloud Platform and respond in the following format: Choices: A. [choice1] B. [choice2] C. [choice3] D. [choice4] Answer: [A/B/C/D]. Please avoid at all cost any comments or explanation!`, question.Text, strings.Join(question.Sections, ","), strings.Join(question.Labels, ","))

	// Using Chat completion
	resp, err := llm.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaudeInstant1Dot2,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages error: %v\n", err)
		}
	}

	// Parse LLM Output.
	llmOutput, err := parseLLMOutput(*resp.Content[0].Text)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	// Parse to MultipleChoiceQuestion
	mQuestion := parseToMultipleQuestion(question, llmOutput.Choices, llmOutput.Answer, llmOutput.Explanation)

	return mQuestion, nil
}

func init() {
	// type arguments [context.Context, *Question, *MultipleChoiceQuestion, error]
	register.DoFn2x2(&LLMClient{})
}

func (llm *LLMClient) setupClient() {
	llm.client = anthropic.NewClient("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")
}

type LLMOutput struct {
	Choices     []Choice `json:"choices"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

func parseLLMOutput(modelOutput string) (*LLMOutput, error) {
	llmOutput := &LLMOutput{}
	fmt.Println(modelOutput)

	// Regular expressions for parsing
	choiceRegex := regexp.MustCompile(`([A-D])\. (.*)`)
	answerRegex := regexp.MustCompile(`Answer: ([A-D])`)
	explanationRegex := regexp.MustCompile(`Explanation: (.*)`)

	// Split input into lines
	lines := strings.Split(modelOutput, "\n")

	// Parse choices
	var currentChoice *Choice
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if match := choiceRegex.FindStringSubmatch(line); match != nil {
			if currentChoice != nil {
				llmOutput.Choices = append(llmOutput.Choices, *currentChoice)
			}
			currentChoice = &Choice{
				Label: match[1],
				Text:  match[2],
			}
		} else if currentChoice != nil {
			currentChoice.Text += " " + line
		}
	}
	if currentChoice != nil {
		llmOutput.Choices = append(llmOutput.Choices, *currentChoice)
	}

	// Parse answer
	if match := answerRegex.FindStringSubmatch(modelOutput); match != nil {
		llmOutput.Answer = match[1]
	}

	// Parse explanation
	if match := explanationRegex.FindStringSubmatch(modelOutput); match != nil {
		llmOutput.Explanation = strings.TrimSpace(match[1])
	}

	return llmOutput, nil
}

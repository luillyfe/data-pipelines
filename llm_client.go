package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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

	prompt := fmt.Sprintf(`Analyze the following question and provide: 1. Four choices that could be used as answers. 2. Indicate which choice is correct. 3. The choices makes the question easy to answer. Question: %s , the question belongs to the following exam's sections: %s and has ben tag with the following labels: %s. Please keep your output scoped to the Google Cloud Platform and respond in the following format: Choices: A. [choice1] B. [choice2] C. [choice3] D. [choice4] Answer: [A/B/C/D]`, question.Text, strings.Join(question.Sections, ","), strings.Join(question.Labels, ","))

	// Using Chat completion
	chatRes, err := llm.client.Chat(llm.ModelName, []mistral.ChatMessage{{Content: prompt, Role: mistral.RoleUser}}, &mistral.ChatRequestParams{
		Temperature: 0.7,
		MaxTokens:   512,
		TopP:        1,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting chat completion: %w", err)
	}

	// Parse parseLLMOutput.
	llmOutput, err := parseLLMOutput(chatRes.Choices[0].Message.Content)
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
	llm.client = mistral.NewMistralClientDefault("")
}

type LLMOutput struct {
	Choices     []Choice `json:"choices"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

func parseLLMOutput(modelOutput string) (*LLMOutput, error) {
	llmOutput := &LLMOutput{}

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

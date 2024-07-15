package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/luillyfe/data-pipelines/llm"
)

// ContextualDataAugmentation is a PTransform that uses an LLM for Contextual Data Augmentation
type ContextualDataAugmentation struct {
	PromptFile string
	model      llm.LanguageModel
	once       sync.Once
}

func (cda *ContextualDataAugmentation) ProcessElement(ctx context.Context, question *Question) (*MultipleChoiceQuestion, error) {
	// Lazy initialization
	cda.once.Do(func() {
		cda.model = llm.NewAnthropicLLM("")
	})

	// Building the prompt
	prompt, err := readFile(cda.PromptFile)
	if err != nil {
		return nil, fmt.Errorf("error in reading the prompt from file: %w", err)
	}
	formattedPrompt := fmt.Sprintf(prompt, question.Text, strings.Join(question.Sections, ","), strings.Join(question.Labels, ","))

	// Using chat completion
	llmOutput, err := cda.model.GenerateText(ctx, formattedPrompt)
	if err != nil {
		return nil, fmt.Errorf("error in generating text from LLM: %w", err)
	}

	// Parse LLM Output.
	ParsedLLMOutput, err := parseLLMOutput(llmOutput)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	// Parse to MultipleChoiceQuestion
	mQuestion := ConvertToMultipleChoiceQuestion(question, ParsedLLMOutput)

	// Return
	return &mQuestion, nil
}

func init() {
	// type arguments [context.Context, *Question, *MultipleChoiceQuestion, error]
	register.DoFn2x2(&ContextualDataAugmentation{})
}

type ParsedDataFromLLM struct {
	ReformulatedQuestion string
	Choices              map[string]string
	CorrectAnswer        string
	Explanation          string
	CoreConcept          string
	DifficultyLevel      string
	QuestionType         string
}

func parseLLMOutput(output string) (*ParsedDataFromLLM, error) {
	lines := strings.Split(output, "\n")
	data := &ParsedDataFromLLM{
		Choices: make(map[string]string),
	}

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "Reformulated Question:"):
			currentSection = "question"
			data.ReformulatedQuestion = strings.TrimPrefix(line, "Reformulated Question:")
		case line == "Choices:":
			currentSection = "choices"
		case strings.HasPrefix(line, "Correct Answer:"):
			data.CorrectAnswer = strings.TrimSpace(strings.TrimPrefix(line, "Correct Answer:"))
		case strings.HasPrefix(line, "Explanation:"):
			currentSection = "explanation"
			data.Explanation = strings.TrimPrefix(line, "Explanation:")
		case strings.HasPrefix(line, "Analysis:"):
			currentSection = "analysis"
		case strings.HasPrefix(line, "- Core Concept:"):
			data.CoreConcept = strings.TrimPrefix(line, "- Core Concept:")
		case strings.HasPrefix(line, "- Difficulty Level:"):
			data.DifficultyLevel = strings.TrimPrefix(line, "- Difficulty Level:")
		case strings.HasPrefix(line, "- Question Type:"):
			data.QuestionType = strings.TrimPrefix(line, "- Question Type:")
		default:
			switch currentSection {
			case "choices":
				parts := strings.SplitN(line, ".", 2)
				if len(parts) == 2 {
					data.Choices[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			case "explanation":
				data.Explanation += " " + line
			}
		}
	}

	// Trim any leading/trailing whitespace
	data.ReformulatedQuestion = strings.TrimSpace(data.ReformulatedQuestion)
	data.Explanation = strings.TrimSpace(data.Explanation)
	data.CoreConcept = strings.TrimSpace(data.CoreConcept)
	data.DifficultyLevel = strings.TrimSpace(data.DifficultyLevel)
	data.QuestionType = strings.TrimSpace(data.QuestionType)

	// Validate that we've extracted all required fields
	if data.ReformulatedQuestion == "" || len(data.Choices) != 4 || data.CorrectAnswer == "" ||
		data.Explanation == "" || data.CoreConcept == "" || data.DifficultyLevel == "" || data.QuestionType == "" {
		return nil, fmt.Errorf("failed to extract all required fields from LLM output")
	}

	return data, nil
}

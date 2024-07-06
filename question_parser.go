package main

import (
	"encoding/json"
	"fmt"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
)

type Question struct {
	Text     string   `json:"text"`
	Type     string   `json:"type"`
	Author   string   `json:"author"`
	Sections []string `json:"sections"`
	Labels   []string `json:"labels"`
}

type Choice struct {
	Label string `json:"option"`
	Text  string `json:"description"`
}

type MultipleChoiceQuestion struct {
	Question    *Question
	Choices     []Choice
	Answer      string
	Explanation string
}

func readQuestions(s beam.Scope, filename string) beam.PCollection {
	s = s.Scope("ReadQuestions")

	// ARD: Read the file at once. We need access to the whole json
	// byte string at once (indented json) to be able to unmarshal it
	// to the proper struct.
	jsonContent := readFile(filename)

	// Create a PCollection
	questionsString := createPCollection(s, jsonContent)

	// Parse a PCollection of []any to a PCollection of []Question
	questions := beam.ParDo(s, func(question string, emit func(*Question)) {
		var q Question
		if err := json.Unmarshal([]byte(question), &q); err != nil {
			fmt.Errorf("Error during Unmarshal(): %w", err)
			return
		}
		emit(&q)
	}, questionsString)

	return questions
}

func createPCollection(s beam.Scope, data string) beam.PCollection {
	s = s.Scope("CreatePCollection")

	// Create a slice to hold the questions
	var questions []Question

	// Unmarshal the JSON data into the slice of Question structs
	err := json.Unmarshal([]byte(data), &questions)
	if err != nil {
		fmt.Errorf("Error during Unmarshal(): %w", err)
	}

	// In order to create a PCollection of individual questions you must cast []Question to []any
	return beam.Create(s, questionsToAny(questions)...)
}

// Function to convert a slice of Question to a slice of any
func questionsToAny(questions []Question) []any {
	var anySlice []any
	for _, q := range questions {
		// Marshal the Question struct to JSON
		jsonBytes, err := json.Marshal(q)
		if err != nil {
			fmt.Errorf("Error marshalling Question to JSON: %w", err)
			continue // Skip to the next question if there's an error
		}

		// Append the JSON string to the anySlice
		anySlice = append(anySlice, string(jsonBytes))
	}
	return anySlice
}

func isNilQuestion(q *Question) bool {
	return q == nil
}

func validateQuestion(q *Question) error {
	if q.Text == "" {
		return fmt.Errorf("question text is empty")
	}
	if q.Type == "" {
		return fmt.Errorf("question type is empty")
	}
	if q.Author == "" {
		return fmt.Errorf("question author is empty")
	}
	return nil
}

func parseToMultipleQuestion(question *Question, choices []Choice, answer, explanation string) *MultipleChoiceQuestion {
	mQuestion := &MultipleChoiceQuestion{
		Question:    question,
		Choices:     choices,
		Answer:      answer,
		Explanation: explanation,
	}

	// Since it has choices, the question's type needs to reflect it
	question.Type = "multiple_choice"

	return mQuestion
}

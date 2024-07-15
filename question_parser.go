package main

import (
	"encoding/json"
	"fmt"
	"time"

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
	Label     string `json:"option"`
	Text      string `json:"description"`
	IsCorrect bool   `json:"is_correct" firestore:"is_correct"`
}

// MultipleChoiceQuestion represents the structure for multiple-choice questions
type MultipleChoiceQuestion struct {
	// Identification: Ensures uniqueness across distributed systems
	ID     string `json:"id" firestore:"id"`         // Unique identifier for the question
	Author string `json:"author" firestore:"author"` // Creator of the question

	// Question Content: Could support adaptive testing
	OriginalText     string   `json:"original_text" firestore:"original_text"`         // Initial wording of the question
	ReformulatedText string   `json:"reformulated_text" firestore:"reformulated_text"` // Alternative or improved wording
	Choices          []Choice `json:"choices" firestore:"choices"`                     // Array of possible answers
	Explanation      string   `json:"explanation" firestore:"explanation"`             // Clarification of the correct answer

	// Categorization: It will help in creating concept maps and learning pathways
	Sections        []string `json:"sections" firestore:"sections"`                 // Topic areas the question belongs to
	Labels          []string `json:"labels" firestore:"labels"`                     // Tags for additional categorization
	CoreConcept     string   `json:"core_concept" firestore:"core_concept"`         // Main idea being tested
	QuestionType    string   `json:"question_type" firestore:"question_type"`       // How challenging the question is
	DifficultyLevel string   `json:"difficulty_level" firestore:"difficulty_level"` // Type of question ("multiple choice")

	// Metadata: It might be use them for audit trails
	CreatedAt   time.Time `json:"created_at" firestore:"created_at"`     // When the question was first created
	UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`     // When the question was last modified
	IsPublished bool      `json:"is_published" firestore:"is_published"` // Whether the question is available for use
}

func readQuestions(s beam.Scope, filename string) beam.PCollection {
	s = s.Scope("ReadQuestions")

	// ARD: Read the file at once. We need access to the whole json
	// byte string at once (indented json) to be able to unmarshal it
	// to the proper struct.
	jsonContent, err := readFile(filename)
	if err != nil {
		return beam.PCollection{}
	}

	// Create a PCollection
	questionsString, err := createPCollection(s, jsonContent)
	if err != nil {
		return beam.PCollection{}
	}

	// Parse a PCollection of []any to a PCollection of []Question
	questions := beam.ParDo(s, func(question string, emit func(*Question)) {
		var q Question
		if err := json.Unmarshal([]byte(question), &q); err != nil {
			fmt.Println(fmt.Errorf("error during Unmarshal(): %w", err))
			return
		}
		emit(&q)
	}, questionsString)

	return questions
}

func createPCollection(s beam.Scope, data string) (beam.PCollection, error) {
	s = s.Scope("CreatePCollection")

	// Create a slice to hold the questions
	var questions []Question

	// Unmarshal the JSON data into the slice of Question structs
	err := json.Unmarshal([]byte(data), &questions)
	if err != nil {
		return beam.PCollection{}, fmt.Errorf("error during Unmarshal(): %w", err)
	}

	// In order to create a PCollection of individual questions you must cast []Question to []any
	return beam.Create(s, questionsToAny(questions)...), nil
}

// Function to convert a slice of Question to a slice of any
func questionsToAny(questions []Question) []any {
	var anySlice []any
	for _, q := range questions {
		// Marshal the Question struct to JSON
		jsonBytes, err := json.Marshal(q)
		if err != nil {
			fmt.Println(fmt.Errorf("error marshalling Question to JSON: %w", err))
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

// Helper function to convert QuestionData to MultipleChoiceQuestion
func ConvertToMultipleChoiceQuestion(originalQuestion *Question, questionData *ParsedDataFromLLM) MultipleChoiceQuestion {
	choices := make([]Choice, 0, len(questionData.Choices))
	for letter, text := range questionData.Choices {
		choices = append(choices, Choice{
			Text:      text,
			IsCorrect: letter == questionData.CorrectAnswer,
		})
	}

	return MultipleChoiceQuestion{
		// ID:               originalQuestion.ID,
		OriginalText:     originalQuestion.Text,
		ReformulatedText: questionData.ReformulatedQuestion,
		Author:           originalQuestion.Author,
		Sections:         originalQuestion.Sections,
		Labels:           originalQuestion.Labels,
		Choices:          choices,
		Explanation:      questionData.Explanation,
		CoreConcept:      questionData.CoreConcept,
		DifficultyLevel:  questionData.DifficultyLevel,
		QuestionType:     questionData.QuestionType,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		IsPublished:      false, // Default to unpublished
	}
}

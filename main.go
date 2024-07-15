package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/filter"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

func main() {
	// Init Apache Beam
	beam.Init()

	// Get env vars
	PROJECT_ID := os.Getenv("PROJECT_ID")
	COLLECTION := os.Getenv("COLLECTION")
	CREDENTIALS_PATH := os.Getenv("CREDENTIALS_PATH")
	PROMPT_FILE := os.Getenv("PROMPT_FILE")

	// Flag for input and output files
	input := flag.String("input", "input.txt", "Input file to process")
	flag.Parse()

	// Create the pipeline
	p := beam.NewPipeline()
	s := p.Root()

	// Read questions as PCollection of []Question
	questions := readQuestions(s, *input)

	// Validate questions to match schema (text, type, author)
	validQuestions := beam.ParDo(s, func(q *Question) (*Question, error) {
		if err := validateQuestion(q); err != nil {
			return nil, err
		}
		return q, nil
	}, questions)

	// Filter out nil questions
	validQuestions = filter.Exclude(s, validQuestions, isNilQuestion)

	// Contextual Data Augmentation (From free-response to Multiple-Choice Questions)
	contextualDataAugmentation := &ContextualDataAugmentation{PromptFile: PROMPT_FILE}
	mQuestions := beam.ParDo(s, contextualDataAugmentation, validQuestions)

	// Initialize the firestore writer
	firestoreWriter := &FirestoreWriter{
		ProjectID:  PROJECT_ID,
		Collection: COLLECTION,
		CredPath:   CREDENTIALS_PATH,
	}

	// Write to Firestore
	beam.ParDo0(s, firestoreWriter, mQuestions)

	// Run the pipeline
	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}

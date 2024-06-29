package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/filter"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

func main() {
	// Init Apache Beam
	beam.Init()

	// Flag for input and output files
	input := flag.String("input", "input.txt", "Input file to process")
	output := flag.String("output", "output.txt", "Output file to write results")
	flag.Parse()

	// Create the pipeline
	p := beam.NewPipeline()
	s := p.Root()

	// Read questions as PCollection of []Question
	questions := readQuestions(s, *input)

	// Validate questions to math schema (id, text)
	validQuestions := beam.ParDo(s, func(q *Question) (*Question, error) {
		if err := validateQuestion(q); err != nil {
			return nil, err
		}
		return q, nil
	}, questions)

	// Filter out nil questions
	validQuestions = filter.Exclude(s, validQuestions, isNilQuestion)

	// Write the processed questions to a JSONL file
	textio.Write(s, *output, beam.ParDo(s, func(q *Question) string {
		data, _ := json.Marshal(q)
		return string(data)
	}, validQuestions))

	// Run the pipeline
	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}

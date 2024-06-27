package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/stats"
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

	// Read lines from input file
	lines := textio.Read(s, *input)

	// Convert lines to lowercase and split into words
	words := beam.ParDo(s, func(line string, emit func(string)) {
		for _, word := range regexp.MustCompile(`\w+`).FindAllString(line, -1) {
			emit(word)
		}
	}, lines)

	// Count word occurrences
	counted := stats.Count(s, words)

	// Filter out words with count less than 5
	filtered := beam.ParDo(s, func(word string, count int) (string, int) {
		if count >= 5 {
			return word, count
		}
		return "", 0
	}, counted)

	// Format results
	formatted := beam.ParDo(s, func(word string, count int) string {
		if word != "" {
			return fmt.Sprintf("%s: %d", word, count)
		}
		return ""
	}, filtered)

	// Remove empty strings
	nonEmpty := beam.ParDo(s, func(s string, emit func(string)) {
		if s != "" {
			emit(s)
		}
	}, formatted)

	// Write results to output file
	textio.Write(s, *output, nonEmpty)

	// Run the pipeline
	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}

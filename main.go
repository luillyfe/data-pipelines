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
	words := beam.ParDo(s, func(line string) []string {
		return regexp.MustCompile(`\w+`).FindAllString(line, -1)
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
		return fmt.Sprintf("%s: %d", word, count)
	}, filtered)

	// Write results to output file
	textio.Write(s, *output, formatted)

	// Run the pipeline
	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}

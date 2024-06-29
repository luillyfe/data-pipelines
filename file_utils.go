package main

import (
	"bufio"
	"fmt"
	"os"
)

func readFile(filename string) string {
	var data string

	file, err := os.Open(filename)
	if err != nil {
		fmt.Errorf("Error when opening file: %w", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Errorf("Error when opening file: %w", err)
		return ""
	}

	return data
}

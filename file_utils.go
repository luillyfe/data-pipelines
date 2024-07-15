package main

import (
	"bufio"
	"fmt"
	"os"
)

func readFile(filename string) (string, error) {
	var data string

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error when opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error when reading file: %w", err)
	}

	return data, nil
}

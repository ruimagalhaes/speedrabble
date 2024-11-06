package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func filter() {
	// Open the original JSON file
	inputFile, err := os.Open("words_dictionary.json")
	if err != nil {
		log.Fatalf("Failed to open input file: %s", err)
	}
	defer inputFile.Close()

	// Read the file content
	byteValue, err := io.ReadAll(inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %s", err)
	}

	// Declare a map to hold the words
	wordMap := make(map[string]int)

	// Unmarshal JSON data into the map
	if err := json.Unmarshal(byteValue, &wordMap); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	// Filter out words longer than 7 letters
	filteredWordMap := make(map[string]int)
	for word, value := range wordMap {
		if len(word) <= 7 {
			filteredWordMap[word] = value
		}
	}

	// Marshal the filtered map back to JSON
	filteredJSON, err := json.MarshalIndent(filteredWordMap, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal filtered JSON: %s", err)
	}

	// Write the filtered JSON to a new file
	outputFile, err := os.Create("filtered_words.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %s", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.Write(filteredJSON); err != nil {
		log.Fatalf("Failed to write to output file: %s", err)
	}

	fmt.Println("Filtered JSON saved to filtered_words.json")
}

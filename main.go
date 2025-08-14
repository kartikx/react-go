package main

import (
	"fmt"
	"encoding/json"
)

func main() {
	fmt.Println("Hello, World!")
	
	// Test the schema generation from tools.go
	fmt.Println("Testing schema generation...")
	
	// Generate and display the schema for ReadFileInput
	toolDefinition := ReadFileDefinition

	fmt.Printf("Tool Definition: %+v\n", toolDefinition)
	
	// Pretty print the schema
	toolDefinitionJson, err := json.MarshalIndent(toolDefinition, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling tool definition: %v\n", err)
	}

	fmt.Printf("Generated Schema:\n%s\n", string(toolDefinitionJson))
}
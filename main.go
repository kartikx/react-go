package main

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()

	// Create a documentation agent
	agent := NewDocAgent(&client)

	// Start the agent's HTTP server
	agent.Start()

	// Run the agent (this will block and handle requests)
	agent.Run(context.Background())
}
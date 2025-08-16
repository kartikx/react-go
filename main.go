package main

import (
	"context"
	"fmt"
	"os"

	"flag"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	agentType := flag.String("agent", "doc", "The type of agent to run. Valid values are 'doc' or 'coder'.")

	flag.Parse()

	client := anthropic.NewClient()

	var agent *Agent

	switch *agentType {
	case "doc":
		agent = NewDocAgent(&client)
	case "coder":
		agent = NewCoderAgent(&client)
	default:
		fmt.Printf("Unknown agent type: %s. Valid values are 'doc' or 'coder'.\n", *agentType)
		os.Exit(1)
	}

	// Start the agent's HTTP server
	agent.Start()

	// Run the agent (this will block and handle requests)
	agent.Run(context.Background())
}
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()

	// TODO - this should move into an agent.
	// Also need to register one agent with another agent. Should go in as a hook.
	getCliInput := func() (string, error) {
		fmt.Printf("Enter input: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		return strings.TrimSpace(input), err
	}

	agent := NewDocAgent(&client, getCliInput)

	agent.Run(context.Background())
}
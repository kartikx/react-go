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

	getCliInput := func() (string, error) {
		fmt.Printf("Enter input: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		return strings.TrimSpace(input), err
	}

	agent := NewAgent(&client, []ToolDefinition{ReadFileDefinition, ExecuteCommandDefinition}, getCliInput)

	agent.Run(context.Background())
}
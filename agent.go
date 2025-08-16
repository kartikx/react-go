package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

type Agent struct {
	client *anthropic.Client
	// This pattern is great, because it can be used to get input from a user, or from
	// another agent.
	getInput func() (string, error)
	tools    []ToolDefinition
}

func NewAgent(client *anthropic.Client, tools []ToolDefinition, getInput func() (string, error)) *Agent {
	return &Agent{
		client:   client,
		tools:    tools,
		getInput: getInput,
	}
}

func (a *Agent) Run(ctx context.Context) (string, error) {
	takeInput := true

	messages := []anthropic.MessageParam{}

	anthropicTools := []anthropic.ToolUnionParam{}

	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name: tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	for {
		if takeInput {
			input, err := a.getInput()
			if err != nil {
				return "", err
			}

			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))
		}

		response, err := a.Infer(ctx, messages, anthropicTools)
		if err != nil {
			return "", err
		}

		messages = append(messages, response.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}

		for _, content := range response.Content {
			switch block := content.AsAny().(type) {
			case anthropic.TextBlock:
				fmt.Printf("Text: %s\n", block.Text)
			case anthropic.ToolUseBlock:
				fmt.Printf("Tool: %s\n", block.Name)
				toolResult := a.ExecuteTool(block.ID, block.Name, block.Input)
				toolResults = append(toolResults, toolResult)
			}
		}

		if len(toolResults) == 0 {
			takeInput = true
		} else {
			takeInput = false
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}
	}
}

func (a *Agent) Infer(ctx context.Context, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) (*anthropic.Message, error) {
	response, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Model: anthropic.ModelClaude3_5Haiku20241022,
		Messages: messages,
		Tools: tools,	
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (a *Agent) ExecuteTool(toolID string, toolName string, toolInput json.RawMessage) (anthropic.ContentBlockParamUnion) {
	fmt.Printf("Executing tool %s with input %s\n", toolName, toolInput)

	time.Sleep(1 * time.Second)

	var toolDef ToolDefinition
	toolFound := false

	for _, tool := range a.tools {
		if tool.Name == toolName {
			toolFound = true
			toolDef = tool
			break
		}
	}

	if !toolFound {
		return anthropic.NewToolResultBlock(toolID, "Tool not found", true)
	}


	// This is the reason why our function takes in a json.RawMessage.
	result, err := toolDef.Function(toolInput)
	if err != nil {
		return anthropic.NewToolResultBlock(toolID, err.Error(), true)
	}

	return anthropic.NewToolResultBlock(toolID, result, false)
}
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

type Agent struct {
	name string
	port int

	client *anthropic.Client
	// This pattern is great, because it can be used to get input from a user, or from
	// another agent.
	readInput func() (string, error)
	writeOutput func(string) error
	tools    []ToolDefinition
	
	// Network request context for channel-based handling
	// TODO - should this be in Agent?
	requestChan chan *http.Request
	responseChan chan http.ResponseWriter
	doneChan chan bool
}

// TODO - ordering of params.
func NewCoderAgent(client *anthropic.Client) *Agent {
	fmt.Println("Creating coder agent")
	agent := NewAgent(client, CoderTools, nil, nil, "coder", 8080)
	
	// Override the read/write functions to work with network context
	agent.readInput = agent.readFromNetwork
	agent.writeOutput = agent.writeToNetwork
	
	return agent
}

func NewDocAgent(client *anthropic.Client) *Agent {
	fmt.Println("Creating doc agent")
	agent := NewAgent(client, DocTools, nil, nil, "doc", 8081)
	
	// Override the read/write functions to work with network context
	// TODO - should this be in Agent? Could I benefit from interfaces here?
	agent.readInput = agent.readFromNetwork
	agent.writeOutput = agent.writeToNetwork
	
	return agent
}

func NewAgent(client *anthropic.Client, tools []ToolDefinition, readInput func() (string, error), writeOutput func(string) error, name string, port int) *Agent {
	return &Agent{
		name: name,
		client:   client,
		tools:    tools,
		readInput: readInput,
		writeOutput: writeOutput,
		port: port,
		requestChan: make(chan *http.Request, 1),
		responseChan: make(chan http.ResponseWriter, 1),
		doneChan: make(chan bool, 1),
	}
}

func (a *Agent) Start() error {
	// Set up HTTP handlers
	http.HandleFunc(fmt.Sprintf("/%s", a.name), a.handleRequest)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s agent is healthy", a.name)))
	})
	
	fmt.Printf("Setting up HTTP handlers for %s agent on port %d\n", a.name, a.port)
	fmt.Printf("Endpoints available: /%s, /health\n", a.name)
	
	// Start the agent on the port.
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", a.port), nil)
		if err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

func (a *Agent) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	fmt.Println("Handling request")
	
	// Send the request to the agent's channel
	a.requestChan <- r
	
	// Store the response writer
	a.responseChan <- w
	
	// Wait for the agent to process and write the response
	// The agent will call writeToNetwork which will write directly to this response writer
	// We need to wait here until the agent is done
	
	// Wait for completion signal
	// TODO - could do error handling here, the channel could take an error.
	<-a.doneChan
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
			input, err := a.readInput()
			if err != nil {
				return "", err
			}

			fmt.Println("Received input: ", input)

			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))
		}

		response, err := a.Infer(ctx, messages, anthropicTools)
		if err != nil {
			return "", err
		}

		messages = append(messages, response.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}

		// fmt.Println("\tReceived response... ")

		ch := make(chan anthropic.ContentBlockParamUnion, 20)

		toolCount := 0	

		for _, content := range response.Content {
			switch block := content.AsAny().(type) {
			case anthropic.TextBlock:
				// fmt.Printf("Text: %s\n", block.Text)
			case anthropic.ToolUseBlock:
				// fmt.Printf("Tool: %s\n", block.Name)
				go a.ExecuteTool(block.ID, block.Name, block.Input, ch)
				toolCount++
				// toolResults = append(toolResults, toolResult)
			}
		}

		for i := 0; i < toolCount; i++ {
			fmt.Printf("Received tool result %d\n", i)
			toolResults = append(toolResults, <-ch)
		}

		if len(toolResults) == 0 {
			takeInput = true
			a.writeOutput(response.Content[0].Text)
		} else {
			takeInput = false
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}
	}
}

func (a *Agent) Infer(ctx context.Context, messages []anthropic.MessageParam, tools []anthropic.ToolUnionParam) (*anthropic.Message, error) {
	response, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		// Model: anthropic.ModelClaude3_5Haiku20241022,
		Model: anthropic.ModelClaudeSonnet4_20250514,
		Messages: messages,
		Tools: tools,
		System: []anthropic.TextBlockParam{
			{
				Text: "<use_parallel_tool_calls> For maximum efficiency, whenever you perform multiple independent operations, invoke all relevant tools simultaneously rather than sequentially. Prioritize calling tools in parallel whenever possible. For example, when reading 3 files, run 3 tool calls in parallel to read all 3 files into context at the same time. When running multiple read-only commands like `ls` or `list_dir`, always run all of the commands in parallel. Err on the side of maximizing parallel tool calls rather than running too many tools sequentially. </use_parallel_tool_calls>",
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (a *Agent) ExecuteTool(toolID string, toolName string, toolInput json.RawMessage, ch chan<- anthropic.ContentBlockParamUnion) {
	fmt.Printf("Executing tool %s with input %s\n", toolName, toolInput)

	// TODO - remove this
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
		ch <- anthropic.NewToolResultBlock(toolID, "Tool not found", true)
		return
	}


	// This is the reason why our function takes in a json.RawMessage.
	result, err := toolDef.Function(toolInput)
	if err != nil {
		fmt.Println("Error executing tool: ", err)
		ch <- anthropic.NewToolResultBlock(toolID, err.Error(), true)
		return
	}

	ch <- anthropic.NewToolResultBlock(toolID, result, false)
}

// readFromNetwork reads input from the stored request context
func (a *Agent) readFromNetwork() (string, error) {
	fmt.Println("Reading from network")

	// Wait for a request to come in
	req := <-a.requestChan
	
	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %v", err)
	}
	
	return string(body), nil
}

// writeToNetwork writes output to the stored response context
func (a *Agent) writeToNetwork(message string) error {
	fmt.Println("Writing to network")


	// Get the response writer
	w := <-a.responseChan
	
	// Write the response
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	
	// Signal completion to the HTTP handler
	a.doneChan <- true
	
	return err
}
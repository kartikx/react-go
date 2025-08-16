package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func readFromCli() (string, error) {
	fmt.Printf("Enter input: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	return strings.TrimSpace(input), err
}

func writeToCli(message string) error {
	fmt.Println(message)
	return nil
}

// Coder-specific tools
var CoderTools = []ToolDefinition{
	ReadFileDefinition,
	ExecuteCommandDefinition,
}

// ReadFile tool for reading file contents
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The path of the file." jsonschema_default:"."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a file. Use this when you want to see what is inside a file.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}

	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// ExecuteCommand tool for running shell commands
type ExecuteCommandInput struct {
	Command string `json:"command" jsonschema_description:"The command to execute"`
}

var ExecuteCommandInputSchema = GenerateSchema[ExecuteCommandInput]()

var ExecuteCommandDefinition = ToolDefinition{
	Name:        "execute_command",
	Description: "Execute a shell command and return the output. Use this when you need to run terminal commands.",
	InputSchema: ExecuteCommandInputSchema,
	Function:    ExecuteCommand,
}

func ExecuteCommand(input json.RawMessage) (string, error) {
	readFileInput := ExecuteCommandInput{}

	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", err
	}
	
	// Split the command into command and arguments
	parts := strings.Fields(readFileInput.Command)
	if len(parts) == 0 {
		return "", nil
	}
	
	cmd := exec.Command(parts[0], parts[1:]...)
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command: %s\nOutput:\n%s", readFileInput.Command, string(output)), err
	}
	
	return fmt.Sprintf("Command: %s\nOutput:\n%s", readFileInput.Command, string(output)), nil
}


// Invoke documentation agent.
type InvokeDocumentationAgentInput struct {
	Query string `json:"query" jsonschema_description:"The query to search for in the documentation"`
}

var InvokeDocumentationAgentInputSchema = GenerateSchema[InvokeDocumentationAgentInput]()

var InvokeDocumentationAgentDefinition = ToolDefinition{
	Name:        "invoke_documentation_agent",
	Description: "Invoke the documentation agent to search for information. Use this when you need to find documentation for a specific package or function.",
	InputSchema: InvokeDocumentationAgentInputSchema,
	Function:    InvokeDocumentationAgent,
}

func InvokeDocumentationAgent(input json.RawMessage) (string, error) {
	invokeDocumentationAgentInput := InvokeDocumentationAgentInput{}

	err := json.Unmarshal(input, &invokeDocumentationAgentInput)
	if err != nil {
		return "", err
	}

	reqBody, err := json.Marshal(map[string]string{
		"query": invokeDocumentationAgentInput.Query,
	})
	if err != nil {
		return "", err
	}

	// TODO - how should we figure out the port? perhaps at startup, agents should register with the main agent.
	resp, err := http.Post("http://localhost:8081", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("documentation agent returned status %d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}


	return string(respBytes), nil
}
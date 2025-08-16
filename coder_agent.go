package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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


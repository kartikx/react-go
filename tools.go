package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

// Anthropic Tool Definition.
type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error) `json:"-"`
}

// Generates InputSchema for a given tool handler function.
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	var reflector = jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a file. Use this when you want to see what is inside a file.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

var ExecuteCommandDefinition = ToolDefinition{
	Name:        "execute_command",
	Description: "Execute a shell command and return the output. Use this when you need to run terminal commands.",
	InputSchema: ExecuteCommandInputSchema,
	Function:    ExecuteCommand,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The path of the file." jsonschema_default:"."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

// TODO - what is a json.RawMessage, and how does it get used?
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

type ExecuteCommandInput struct {
	Command string `json:"command" jsonschema_description:"The command to execute"`
}

var ExecuteCommandInputSchema = GenerateSchema[ExecuteCommandInput]()

func ExecuteCommand(input json.RawMessage) (string, error) {
	executeCommandInput := ExecuteCommandInput{}

	err := json.Unmarshal(input, &executeCommandInput)
	if err != nil {
		return "", err
	}
	
	// Split the command into command and arguments
	parts := strings.Fields(executeCommandInput.Command)
	if len(parts) == 0 {
		return "", nil
	}
	
	cmd := exec.Command(parts[0], parts[1:]...)
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command: %s\nOutput:\n%s", executeCommandInput.Command, string(output)), err
	}
	
	return fmt.Sprintf("Command: %s\nOutput:\n%s", executeCommandInput.Command, string(output)), nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// Documentation-specific tools
var DocTools = []ToolDefinition{
	SearchGoDocumentationDefinition,
	SearchRagDatabaseDefinition,
}

// SearchGoDocumentation tool for searching Go documentation
type SearchGoDocumentationInput struct {
	PackageName string `json:"package_name" jsonschema_description:"The name of the package to search for"`
}

var SearchGoDocumentationInputSchema = GenerateSchema[SearchGoDocumentationInput]()

var SearchGoDocumentationDefinition = ToolDefinition{
	Name:        "search_go_documentation",
	Description: "Search Go documentation for information. Use this when you need to find Go language features, standard library functions, or Go-specific information. Call this function with the name of the package you want to search for.",
	InputSchema: SearchGoDocumentationInputSchema,
	Function:    SearchGoDocumentation,
}

// SearchGoDocumentation fetches documentation text from pkg.go.dev for a given package
func SearchGoDocumentation(input json.RawMessage) (string, error) {
	searchInput := SearchGoDocumentationInput{}

	err := json.Unmarshal(input, &searchInput)
	if err != nil {
		return "", err
	}

    url := fmt.Sprintf("https://pkg.go.dev/%s?tab=doc", searchInput.PackageName)
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("failed to fetch package docs: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("failed to fetch package docs: status %d", resp.StatusCode)
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to parse HTML: %v", err)
    }

    // Select the documentation section
    docText := ""
    docSelection := doc.Find(".Documentation-overview")
    if docSelection.Length() == 0 {
        return "", fmt.Errorf("documentation section not found")
    }

    docSelection.Each(func(i int, s *goquery.Selection) {
        docText += s.Text() + "\n"
    })

    return docText, nil
}

// SearchRagDatabase tool for searching the RAG knowledge base
type SearchRagDatabaseInput struct {
	Query string `json:"query" jsonschema_description:"The search query to execute on the RAG database"`
}

var SearchRagDatabaseInputSchema = GenerateSchema[SearchRagDatabaseInput]()

var SearchRagDatabaseDefinition = ToolDefinition{
	Name:        "search_rag_database",
	Description: "Search the RAG (Retrieval-Augmented Generation) database for relevant information. Use this when you need to find specific knowledge from your organization's documentation or knowledge base.",
	InputSchema: SearchRagDatabaseInputSchema,
	Function:    SearchRagDatabase,
}

func SearchRagDatabase(input json.RawMessage) (string, error) {
	searchInput := SearchRagDatabaseInput{}

	err := json.Unmarshal(input, &searchInput)
	if err != nil {
		return "", err
	}

	// TODO: Implement actual RAG database search
	// For now, return a stub response
	return fmt.Sprintf("RAG database search results for: %s\n\n[This is a stub implementation. Actual RAG database integration would be implemented here.]", searchInput.Query), nil
}
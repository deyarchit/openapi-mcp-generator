package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	ctx := context.Background()
	mcpClient, err := client.NewStdioMCPClient(
		"go",
		[]string{},
		"run",
		".",
	)
	assert.NoError(t, err)

	mcpClient.Initialize(ctx, mcp.InitializeRequest{})
	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	assert.NoError(t, err)

	for _, tool := range tools.Tools {
		// Marshal the struct to JSON
		jsonBytes, err := json.MarshalIndent(tool.InputSchema, "", "    ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		// Convert the byte slice to a string
		jsonString := string(jsonBytes)
		fmt.Printf("\nName: %s\n, Description: %s\n, Input Schema: %v", tool.Name, tool.Description, jsonString)
	}

}

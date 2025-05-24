package main

import (
	"context"
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
		fmt.Printf("Name: %s\n, Description: %s\n, OutputSchema: %v", tool.Name, tool.Description, tool.InputSchema)
	}

}

package openapimcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

type ServerMode string

const (
	StdIO ServerMode = "stdio"
	SSE   ServerMode = "sse"
)

// GeneratorConfig holds the configuration for generating and running the MCP server.
type GeneratorConfig struct {
	SpecSource string     // URL or file path to the OpenAPI spec
	ServerMode ServerMode // server.ModeStdIO or server.ModeSSE
}

// RunFromSpec loads an OpenAPI spec, builds an MCP server, and starts it.
func RunFromSpec(config GeneratorConfig) error {
	if config.SpecSource == "" {
		return fmt.Errorf("spec source cannot be empty")
	}

	// 1. Load the OpenAPI spec
	openapiSpec, err := LoadSpec(config.SpecSource)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// 2. Build the MCP server from the spec
	mcpServer, err := BuildMCPServerFromSpec(openapiSpec)
	if err != nil {
		return fmt.Errorf("failed to build MCP server from spec: %w", err)
	}

	// 3. Start the MCP server
	// log.Printf("Starting MCP server '%s' in %s mode...", mcpServer.Name(), mcpServer.Mode())
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server failed to start or exited with error: %w", err)
	}

	// log.Printf("MCP server '%s' has stopped.", mcpServer.Name)
	return nil

}

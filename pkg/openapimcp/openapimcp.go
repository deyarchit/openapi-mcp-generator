package openapimcp

import (
	"fmt"
	"net/http"
	"net/url"

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

	// Load the OpenAPI spec
	openapiSpec, err := LoadSpec(config.SpecSource)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	serverCfg := &APIConfig{
		BaseURL:    getBaseURLFromSpecSource(config.SpecSource),
		HTTPClient: &http.Client{},
		Headers:    make(map[string]string),
	}

	// Build the MCP server from the spec and config
	mcpServer, err := BuildMCPServerFromSpec(openapiSpec, serverCfg)
	if err != nil {
		return fmt.Errorf("failed to build MCP server from spec: %w", err)
	}

	// Start the MCP server
	// log.Printf("Starting MCP server '%s' in %s mode...", mcpServer.Name(), mcpServer.Mode())
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("MCP server failed to start or exited with error: %w", err)
	}

	// log.Printf("MCP server '%s' has stopped.", mcpServer.Name)
	return nil

}

func getBaseURLFromSpecSource(specSource string) string {
	var baseURL string
	u, urlErr := url.ParseRequestURI(specSource)
	if urlErr == nil && (u.Scheme == "http" || u.Scheme == "https") {
		baseURL = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}
	return baseURL
}

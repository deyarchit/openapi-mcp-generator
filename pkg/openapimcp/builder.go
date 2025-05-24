package openapimcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// APIConfig holds configuration for API calls
type APIConfig struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// MCPServerBuilder builds MCP servers from OpenAPI specs
type MCPServerBuilder struct {
	config *APIConfig
}

// NewMCPServerBuilder creates a new builder with configuration
func NewMCPServerBuilder(config *APIConfig) *MCPServerBuilder {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{}
	}
	return &MCPServerBuilder{config: config}
}

// BuildMCPServerFromSpec creates an MCP server from an OpenAPI spec
func (b *MCPServerBuilder) BuildMCPServerFromSpec(spec *openapi3.T) (*server.MCPServer, error) {
	mcpServer := server.NewMCPServer("openapi-server", "1.0.0")

	// Extract base URL from spec if not provided in config
	baseURL := b.config.BaseURL
	if baseURL == "" && len(spec.Servers) > 0 {
		baseURL = spec.Servers[0].URL
	}

	// Process all paths and operations
	for path, pathItem := range spec.Paths.Map() {
		operations := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			toolName := generateToolName(method, path, operation)
			tool := b.createTool(toolName, operation)

			// Create handler for this specific endpoint
			handler := b.createHandler(method, baseURL+path, operation)

			// Register tool with handler
			mcpServer.AddTool(tool, handler)
		}
	}

	return mcpServer, nil
}

// createTool creates an MCP tool from an OpenAPI operation
func (b *MCPServerBuilder) createTool(toolName string, op *openapi3.Operation) mcp.Tool {
	tool := mcp.Tool{
		Name:        toolName,
		Description: op.Description,
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: make(map[string]any),
			Required:   []string{},
		},
	}

	if op.Description != "" && tool.Description == "" {
		tool.Description = op.Description
	}

	// Add parameters to tool schema
	for _, paramRef := range op.Parameters {
		param := paramRef.Value
		if param == nil {
			continue
		}

		paramSchema := convertSchemaToMCP(param.Schema)
		tool.InputSchema.Properties[param.Name] = map[string]any{
			"type":        paramSchema["type"],
			"description": param.Description,
		}

		// Add format, enum, etc. if present
		if format, ok := paramSchema["format"]; ok {
			tool.InputSchema.Properties[param.Name].(map[string]any)["format"] = format
		}
		if enum, ok := paramSchema["enum"]; ok {
			tool.InputSchema.Properties[param.Name].(map[string]any)["enum"] = enum
		}

		if param.Required {
			tool.InputSchema.Required = append(tool.InputSchema.Required, param.Name)
		}
	}

	// Add request body parameters if present
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for mediaType, mediaTypeObj := range op.RequestBody.Value.Content {
			if strings.Contains(mediaType, "json") && mediaTypeObj.Schema != nil {
				bodySchema := convertSchemaToMCP(mediaTypeObj.Schema)
				tool.InputSchema.Properties["requestBody"] = bodySchema
				if op.RequestBody.Value.Required {
					tool.InputSchema.Required = append(tool.InputSchema.Required, "requestBody")
				}
				break
			}
		}
	}

	return tool
}

// createHandler creates a handler function for an API endpoint
func (b *MCPServerBuilder) createHandler(method, fullURL string, op *openapi3.Operation) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse arguments
		args := request.GetArguments()

		// Build URL with path parameters
		finalURL := fullURL
		queryParams := url.Values{}
		var requestBody any

		// Process parameters
		for _, paramRef := range op.Parameters {
			param := paramRef.Value
			if param == nil {
				continue
			}

			value, exists := args[param.Name]
			if !exists && param.Required {
				return nil, fmt.Errorf("required parameter %s is missing", param.Name)
			}

			if !exists {
				continue
			}

			valueStr := fmt.Sprintf("%v", value)

			switch param.In {
			case "path":
				finalURL = strings.ReplaceAll(finalURL, "{"+param.Name+"}", valueStr)
			case "query":
				queryParams.Add(param.Name, valueStr)
			case "header":
				// Headers will be handled separately
			}
		}

		// Add query parameters to URL
		if len(queryParams) > 0 {
			separator := "?"
			if strings.Contains(finalURL, "?") {
				separator = "&"
			}
			finalURL += separator + queryParams.Encode()
		}

		// Handle request body
		if body, exists := args["requestBody"]; exists {
			requestBody = body
		}

		// Make HTTP request
		response, err := b.makeHTTPRequest(ctx, method, finalURL, requestBody, op, args)
		if err != nil {
			return nil, fmt.Errorf("API request failed: %w", err)
		}

		return mcp.NewToolResultText(response), nil
	}
}

// makeHTTPRequest performs the actual HTTP request
func (b *MCPServerBuilder) makeHTTPRequest(ctx context.Context, method, url string, body any, op *openapi3.Operation, args map[string]any) (string, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set configured headers
	for key, value := range b.config.Headers {
		req.Header.Set(key, value)
	}

	// Set parameter headers
	for _, paramRef := range op.Parameters {
		param := paramRef.Value
		if param == nil || param.In != "header" {
			continue
		}

		if value, exists := args[param.Name]; exists {
			req.Header.Set(param.Name, fmt.Sprintf("%v", value))
		}
	}

	resp, err := b.config.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Format response with status information
	result := fmt.Sprintf("Status: %d %s\n", resp.StatusCode, resp.Status)

	if len(responseBody) > 0 {
		// Try to pretty-print JSON
		var jsonObj any
		if err := json.Unmarshal(responseBody, &jsonObj); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				result += "Response:\n" + string(prettyJSON)
			} else {
				result += "Response:\n" + string(responseBody)
			}
		} else {
			result += "Response:\n" + string(responseBody)
		}
	}

	return result, nil
}

// Helper functions

// generateToolName creates a unique tool name from method, path, and operation
func generateToolName(method, path string, op *openapi3.Operation) string {
	if op.OperationID != "" {
		return op.OperationID
	}

	// Clean path for tool name
	cleanPath := regexp.MustCompile(`[{}]`).ReplaceAllString(path, "")
	cleanPath = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(cleanPath, "_")
	cleanPath = strings.Trim(cleanPath, "_")

	return strings.ToLower(method) + "_" + cleanPath
}

// convertSchemaToMCP converts OpenAPI schema to MCP tool schema format
func convertSchemaToMCP(schemaRef *openapi3.SchemaRef) map[string]any {
	return convertSchemaToMCPWithRefs(schemaRef, make(map[string]bool))
}

// convertSchemaToMCPWithRefs converts OpenAPI schema to MCP tool schema format with reference tracking
func convertSchemaToMCPWithRefs(schemaRef *openapi3.SchemaRef, visited map[string]bool) map[string]any {
	if schemaRef == nil {
		return map[string]any{"type": "string"}
	}

	// Handle references to custom models
	if schemaRef.Ref != "" {
		// Check for circular references
		if visited[schemaRef.Ref] {
			// Return a simplified schema for circular references
			return map[string]any{
				"type":        "object",
				"description": fmt.Sprintf("Reference to %s (circular reference detected)", extractRefName(schemaRef.Ref)),
			}
		}

		// Mark this reference as visited
		visited[schemaRef.Ref] = true
		defer func() { delete(visited, schemaRef.Ref) }()

		// If we have the resolved schema, process it
		if schemaRef.Value != nil {
			result := convertSchemaToMCPWithRefs(&openapi3.SchemaRef{Value: schemaRef.Value}, visited)
			// Add reference information
			refName := extractRefName(schemaRef.Ref)
			if desc, ok := result["description"].(string); ok {
				result["description"] = fmt.Sprintf("%s (Model: %s)", desc, refName)
			} else {
				result["description"] = fmt.Sprintf("Model: %s", refName)
			}
			return result
		}

		// If we don't have the resolved schema, return a placeholder
		return map[string]any{
			"type":        "object",
			"description": fmt.Sprintf("Reference to model: %s", extractRefName(schemaRef.Ref)),
		}
	}

	if schemaRef.Value == nil {
		return map[string]any{"type": "string"}
	}

	schema := schemaRef.Value
	result := make(map[string]any)

	// Handle type - schema.Type is *openapi3.Types (slice of strings)
	if schema.Type != nil && len(*schema.Type) > 0 {
		// Use the first type if multiple types are specified
		result["type"] = (*schema.Type)[0]
	} else {
		result["type"] = "string"
	}

	// Get the primary type for comparison
	primaryType := ""
	if schema.Type != nil && len(*schema.Type) > 0 {
		primaryType = (*schema.Type)[0]
	}

	// Handle description
	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Handle format
	if schema.Format != "" {
		result["format"] = schema.Format
	}

	// Handle enum
	if len(schema.Enum) > 0 {
		result["enum"] = schema.Enum
	}

	// Handle object properties
	if primaryType == "object" && len(schema.Properties) > 0 {
		properties := make(map[string]any)
		for propName, propSchema := range schema.Properties {
			properties[propName] = convertSchemaToMCPWithRefs(propSchema, visited)
		}
		result["properties"] = properties

		if len(schema.Required) > 0 {
			result["required"] = schema.Required
		}
	}

	// Handle array items
	if primaryType == "array" && schema.Items != nil {
		result["items"] = convertSchemaToMCPWithRefs(schema.Items, visited)
	}

	// Handle allOf, oneOf, anyOf
	if len(schema.AllOf) > 0 {
		result["allOf"] = make([]any, len(schema.AllOf))
		for i, subSchema := range schema.AllOf {
			result["allOf"].([]any)[i] = convertSchemaToMCPWithRefs(subSchema, visited)
		}
	}

	if len(schema.OneOf) > 0 {
		result["oneOf"] = make([]any, len(schema.OneOf))
		for i, subSchema := range schema.OneOf {
			result["oneOf"].([]any)[i] = convertSchemaToMCPWithRefs(subSchema, visited)
		}
	}

	if len(schema.AnyOf) > 0 {
		result["anyOf"] = make([]any, len(schema.AnyOf))
		for i, subSchema := range schema.AnyOf {
			result["anyOf"].([]any)[i] = convertSchemaToMCPWithRefs(subSchema, visited)
		}
	}

	// Handle number constraints
	if schema.Min != nil {
		result["minimum"] = *schema.Min
	}
	if schema.Max != nil {
		result["maximum"] = *schema.Max
	}

	// Handle string constraints
	if schema.MinLength > 0 {
		result["minLength"] = schema.MinLength
	}
	if schema.MaxLength != nil {
		result["maxLength"] = *schema.MaxLength
	}

	// Handle pattern
	if schema.Pattern != "" {
		result["pattern"] = schema.Pattern
	}

	// Handle array constraints
	if schema.MinItems > 0 {
		result["minItems"] = schema.MinItems
	}
	if schema.MaxItems != nil {
		result["maxItems"] = *schema.MaxItems
	}

	// Handle object constraints
	if schema.MinProps > 0 {
		result["minProperties"] = schema.MinProps
	}
	if schema.MaxProps != nil {
		result["maxProperties"] = *schema.MaxProps
	}

	// Handle default values
	if schema.Default != nil {
		result["default"] = schema.Default
	}

	// Handle examples
	if schema.Example != nil {
		result["example"] = schema.Example
	}

	return result
}

// extractRefName extracts the model name from a reference string
func extractRefName(ref string) string {
	// Handle common reference formats:
	// #/components/schemas/ModelName -> ModelName
	// #/definitions/ModelName -> ModelName
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// Convenience function that matches your requested signature
func BuildMCPServerFromSpec(spec *openapi3.T) (*server.MCPServer, error) {
	config := &APIConfig{
		HTTPClient: &http.Client{},
		Headers:    make(map[string]string),
	}
	builder := NewMCPServerBuilder(config)
	return builder.BuildMCPServerFromSpec(spec)
}

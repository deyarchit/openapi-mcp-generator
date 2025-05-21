package openapimcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetAPIHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(fmt.Sprintf("Echo: %v", request.GetArguments())), nil
	}
}

// // CreateMCPToolHandler creates an mcp.ToolHandlerFunc for a given OpenAPI operation.
// // This handler will process tool calls, log information, and return a structured result.
// func CreateMCPToolHandler(path string, httpMethod string, operation *openapi3.Operation) server.ToolHandlerFunc {
// 	// This ID is primarily for logging and internal reference within the handler.
// 	// The actual tool name registered with MCP will be determined in builder.go.
// 	handlerGeneratedID := operation.OperationID
// 	if handlerGeneratedID == "" {
// 		sanitizedPath := strings.ReplaceAll(strings.Trim(path, "/"), "/", "_")
// 		sanitizedPath = strings.ReplaceAll(sanitizedPath, "{", "")
// 		sanitizedPath = strings.ReplaceAll(sanitizedPath, "}", "")
// 		handlerGeneratedID = fmt.Sprintf("%s_%s", strings.ToUpper(httpMethod), sanitizedPath)
// 	}

// 	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

// 		// The 'request.Name' is the actual tool name this handler was invoked for.
// 		log.Printf("[MCP ToolHandler - %s (OpenAPI: %s %s)] Received call for tool: '%s'",
// 			handlerGeneratedID, strings.ToUpper(httpMethod), path, request.Name)

// 		// Log incoming arguments (which are expected to be a JSON string)
// 		if request.Arguments != "" {
// 			log.Printf("[MCP ToolHandler - %s] Request Arguments (JSON string): %s", handlerGeneratedID, request.Arguments)
// 		} else {
// 			log.Printf("[MCP ToolHandler - %s] Request Arguments: (empty)", handlerGeneratedID)
// 		}

// 		// Attempt to parse arguments string as JSON map for easier inspection/use
// 		var parsedArgs map[string]interface{}
// 		if request.Arguments != "" {
// 			err := json.Unmarshal([]byte(request.Arguments), &parsedArgs)
// 			if err != nil {
// 				log.Printf("[MCP ToolHandler - %s] Warning: Could not unmarshal request arguments as JSON: %v. Treating as raw string.", handlerGeneratedID, err)
// 				// If not valid JSON, we can still include the raw string in the response
// 				parsedArgs = map[string]interface{}{"rawArguments": request.Arguments}
// 			}
// 		} else {
// 			parsedArgs = make(map[string]interface{}) // No arguments provided
// 		}

// 		// Prepare a structured JSON response content
// 		responseContentMap := map[string]interface{}{
// 			"status":                      "success",
// 			"message":                     fmt.Sprintf("Tool '%s' (mapped from OpenAPI: %s %s, handler ID: '%s') processed.", request.Name, strings.ToUpper(httpMethod), path, handlerGeneratedID),
// 			"receivedArguments":           parsedArgs,
// 			"notes":                       "This is a stubbed response from the openapimcp-generator using mcp.ToolHandlerFunc.",
// 			"originalOpenApiOpId":         operation.OperationID, // Original OperationID from spec, if any
// 			"handlerProcessingPath":       path,
// 			"handlerProcessingHttpMethod": httpMethod,
// 		}

// 		responseContentBytes, err := json.Marshal(responseContentMap)
// 		if err != nil {
// 			log.Printf("[MCP ToolHandler - %s] Error marshalling response content: %v", handlerGeneratedID, err)
// 			// Return an error to the MCP framework if we can't form our JSON response
// 			// The MCP framework will then decide how to signal this error.
// 			return nil, fmt.Errorf("failed to marshal response content for tool %s (handler %s): %w", request.Name, handlerGeneratedID, err)
// 		}

// 		log.Printf("[MCP ToolHandler - %s] Sending result content: %s", handlerGeneratedID, string(responseContentBytes))

// 		// Construct the CallToolResult
// 		// The Name in CallToolResult should match the Name in CallToolRequest.
// 		return &mcp.CallToolResult{
// 			Name:    request.Name,
// 			Content: string(responseContentBytes),
// 		}, nil
// 	}
// }

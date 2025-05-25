package openapimcp

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"input_schema"`
}

type ToolInputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}

func TestCreateTool_BasicOperation(t *testing.T) {
	builder := &MCPServerBuilder{}

	op := &openapi3.Operation{
		Description: "Test operation",
	}

	tool := builder.createTool("test_tool", op)

	// Verify OpenAI function calling schema compatibility
	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "Test operation", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.NotNil(t, tool.InputSchema.Properties)
	assert.NotNil(t, tool.InputSchema.Required)
}

func TestCreateTool_WithParameters(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}
	integerType := openapi3.Types{"integer"}

	// Create parameter with schema
	param1 := &openapi3.Parameter{
		Name:        "user_id",
		Description: "User identifier",
		Required:    true,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:   &stringType,
				Format: "uuid",
			},
		},
	}

	param2 := &openapi3.Parameter{
		Name:        "limit",
		Description: "Number of items to return",
		Required:    false,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    &integerType,
				Example: 10,
			},
		},
	}

	op := &openapi3.Operation{
		Description: "Get user data",
		Parameters: []*openapi3.ParameterRef{
			{Value: param1},
			{Value: param2},
		},
	}

	tool := builder.createTool("get_user", op)

	// Verify schema structure matches OpenAI format
	assert.Equal(t, "get_user", tool.Name)
	assert.Equal(t, "Get user data", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)

	// Check properties
	require.Contains(t, tool.InputSchema.Properties, "user_id")
	require.Contains(t, tool.InputSchema.Properties, "limit")

	userIdProp := tool.InputSchema.Properties["user_id"].(map[string]any)
	assert.Equal(t, "string", userIdProp["type"])
	assert.Equal(t, "uuid", userIdProp["format"])
	assert.Equal(t, "User identifier", userIdProp["description"])

	limitProp := tool.InputSchema.Properties["limit"].(map[string]any)
	assert.Equal(t, "integer", limitProp["type"])
	assert.Equal(t, "Number of items to return", limitProp["description"])
	assert.Equal(t, 10, limitProp["example"])

	// Check required fields
	assert.Contains(t, tool.InputSchema.Required, "user_id")
	assert.NotContains(t, tool.InputSchema.Required, "limit")
}

func TestCreateTool_WithEnumParameter(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}

	param := &openapi3.Parameter{
		Name:        "status",
		Description: "Status filter",
		Required:    true,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &stringType,
				Enum: []any{"active", "inactive", "pending"},
			},
		},
	}

	op := &openapi3.Operation{
		Description: "Filter by status",
		Parameters:  []*openapi3.ParameterRef{{Value: param}},
	}

	tool := builder.createTool("filter_status", op)

	statusProp := tool.InputSchema.Properties["status"].(map[string]any)
	assert.Equal(t, "string", statusProp["type"])
	assert.Equal(t, []any{"active", "inactive", "pending"}, statusProp["enum"])
	assert.Equal(t, "Status filter", statusProp["description"])
}

func TestCreateTool_WithRequestBody(t *testing.T) {
	builder := &MCPServerBuilder{}

	objectType := openapi3.Types{"object"}
	stringType := openapi3.Types{"string"}

	// Create request body schema
	requestBodySchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &objectType,
			Properties: map[string]*openapi3.SchemaRef{
				"name": {
					Value: &openapi3.Schema{
						Type:        &stringType,
						Description: "User name",
					},
				},
				"email": {
					Value: &openapi3.Schema{
						Type:   &stringType,
						Format: "email",
					},
				},
			},
			Required: []string{"name"},
		},
	}

	requestBody := &openapi3.RequestBody{
		Content: map[string]*openapi3.MediaType{
			"application/json": {
				Schema: requestBodySchema,
			},
		},
	}

	op := &openapi3.Operation{
		Description: "Create user",
		RequestBody: &openapi3.RequestBodyRef{Value: requestBody},
	}

	tool := builder.createTool("create_user", op)

	// Verify request body properties are included
	require.Contains(t, tool.InputSchema.Properties, "name")
	require.Contains(t, tool.InputSchema.Properties, "email")

	nameProp := tool.InputSchema.Properties["name"].(map[string]any)
	assert.Equal(t, "string", nameProp["type"])

	emailProp := tool.InputSchema.Properties["email"].(map[string]any)
	assert.Equal(t, "string", emailProp["type"])
	assert.Equal(t, "email", emailProp["format"])

	// Verify required fields from request body
	assert.Contains(t, tool.InputSchema.Required, "name")
	assert.NotContains(t, tool.InputSchema.Required, "email")
}

func TestCreateTool_CombinedParametersAndRequestBody(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}
	objectType := openapi3.Types{"object"}

	// Parameter
	param := &openapi3.Parameter{
		Name:     "id",
		Required: true,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{Type: &stringType},
		},
	}

	// Request body
	requestBodySchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &objectType,
			Properties: map[string]*openapi3.SchemaRef{
				"data": {
					Value: &openapi3.Schema{
						Type: &objectType,
					},
				},
			},
			Required: []string{"data"},
		},
	}

	requestBody := &openapi3.RequestBody{
		Content: map[string]*openapi3.MediaType{
			"application/json": {Schema: requestBodySchema},
		},
	}

	op := &openapi3.Operation{
		Description: "Update resource",
		Parameters:  []*openapi3.ParameterRef{{Value: param}},
		RequestBody: &openapi3.RequestBodyRef{Value: requestBody},
	}

	tool := builder.createTool("update_resource", op)

	// Should have both parameter and request body properties
	assert.Contains(t, tool.InputSchema.Properties, "id")
	assert.Contains(t, tool.InputSchema.Properties, "data")

	// Both should be required
	assert.Contains(t, tool.InputSchema.Required, "id")
	assert.Contains(t, tool.InputSchema.Required, "data")
	assert.Len(t, tool.InputSchema.Required, 2)
}

func TestCreateTool_NilParameterHandling(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}

	op := &openapi3.Operation{
		Description: "Test with nil parameter",
		Parameters: []*openapi3.ParameterRef{
			{Value: nil}, // Nil parameter should be skipped
			{Value: &openapi3.Parameter{
				Name: "valid_param",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &stringType},
				},
			}},
		},
	}

	tool := builder.createTool("test_nil", op)

	// Should only contain the valid parameter
	assert.Len(t, tool.InputSchema.Properties, 1)
	assert.Contains(t, tool.InputSchema.Properties, "valid_param")
}

func TestCreateTool_NonJSONRequestBodyIgnored(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}
	objectType := openapi3.Types{"object"}

	requestBody := &openapi3.RequestBody{
		Content: map[string]*openapi3.MediaType{
			"text/plain": {
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &stringType},
				},
			},
			"application/xml": {
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &objectType},
				},
			},
		},
	}

	op := &openapi3.Operation{
		Description: "Non-JSON body",
		RequestBody: &openapi3.RequestBodyRef{Value: requestBody},
	}

	tool := builder.createTool("non_json", op)

	// Should have empty properties since no JSON content
	assert.Len(t, tool.InputSchema.Properties, 0)
	assert.Len(t, tool.InputSchema.Required, 0)
}

func TestCreateTool_OpenAICompatibilitySchema(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}

	param := &openapi3.Parameter{
		Name:        "query",
		Description: "Search query",
		Required:    true,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    &stringType,
				Example: "example search",
			},
		},
	}

	op := &openapi3.Operation{
		Description: "Search function",
		Parameters:  []*openapi3.ParameterRef{{Value: param}},
	}

	tool := builder.createTool("search", op)

	// Verify the schema matches OpenAI function calling format exactly
	assert.Equal(t, "search", tool.Name)
	assert.Equal(t, "Search function", tool.Description)

	// Input schema should be object type with properties and required fields
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.IsType(t, map[string]any{}, tool.InputSchema.Properties)
	assert.IsType(t, []string{}, tool.InputSchema.Required)

	// Property should have correct structure
	queryProp, exists := tool.InputSchema.Properties["query"]
	require.True(t, exists)

	queryPropMap := queryProp.(map[string]any)
	assert.Equal(t, "string", queryPropMap["type"])
	assert.Equal(t, "Search query", queryPropMap["description"])
	assert.Equal(t, "example search", queryPropMap["example"])

	// Required array should contain the required parameter
	assert.Equal(t, []string{"query"}, tool.InputSchema.Required)
}

func TestCreateTool_EmptyRequiredArray(t *testing.T) {
	builder := &MCPServerBuilder{}

	stringType := openapi3.Types{"string"}

	param := &openapi3.Parameter{
		Name:     "optional_param",
		Required: false, // Not required
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{Type: &stringType},
		},
	}

	op := &openapi3.Operation{
		Description: "Optional params only",
		Parameters:  []*openapi3.ParameterRef{{Value: param}},
	}

	tool := builder.createTool("optional_only", op)

	// Required array should be empty but not nil (OpenAI compatibility)
	assert.NotNil(t, tool.InputSchema.Required)
	assert.Len(t, tool.InputSchema.Required, 0)
	assert.Equal(t, []string{}, tool.InputSchema.Required)
}

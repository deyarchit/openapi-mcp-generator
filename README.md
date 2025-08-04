# OpenAPI MCP Generator

A utility for serving api service with openapi v3 spec as a mcp server. 

## Running

`go run github.com/deyarchit/openapi-mcp-generator/cmd/mcp-server-cli@main --spec-file=<open_api_spec_json>`

Add the config to the mcp config file `~/.mcp_config`:
```
{
  "mcpServers": {
    "demo-server": {
        "command": "go",
        "args": [
          "run",
          "github.com/deyarchit/openapi-mcp-generator/cmd/mcp-server-cli@main",
            "--spec-file=<open_api_spec_json>"
        ] 
    }
  }
}
```

```bash
▶ npx @modelcontextprotocol/inspector --config ~/.mcp_config --server <config_dict_name>
Starting MCP inspector...
⚙️ Proxy server listening on localhost:6277
.
.
.
🌐 Opening browser...
New STDIO connection request
.
.
.
STDIO transport: command=/opt/homebrew/opt/go@1.23/bin/go, args=run,github.com/deyarchit/openapi-mcp-generator/cmd/mcp-server-cli@main,--spec-file=<open_api_spec_json>
Created server transport
Created client transport
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b
Received POST message for sessionId 4c899e13-8c00-4b38-9e8e-62b4933ad87b

```



package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"

	"github.com/deyarchit/openapi-mcp-generator/pkg/openapimcp"
)

var (
	specFile string
	specURL  string
	modeStr  string
)

func init() {
	pflag.StringVarP(&specFile, "spec-file", "f", "", "Path to a local OpenAPI spec file (JSON or YAML).")
	pflag.StringVarP(&specURL, "spec-url", "u", "", "URL to a remote OpenAPI spec file (JSON or YAML).")
	pflag.StringVarP(&modeStr, "mode", "m", "stdio", "MCP server mode: 'stdio' or 'sse'. (default: stdio)")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generates and runs an MCP server based on an OpenAPI specification.\n\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -f ./path/to/your/openapi.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -u https://petstore3.swagger.io/api/v3/openapi.json --mode sse\n", os.Args[0])
	}
}

func main() {
	pflag.Parse()

	// Validate flags
	if specFile == "" && specURL == "" {
		log.Println("Error: You must provide either --spec-file or --spec-url.")
		pflag.Usage()
		os.Exit(1)
	}
	if specFile != "" && specURL != "" {
		log.Println("Error: You can only provide one of --spec-file or --spec-url, not both.")
		pflag.Usage()
		os.Exit(1)
	}

	specSource := specFile
	if specURL != "" {
		specSource = specURL
	}

	// var mcpMode openapimcp.ServerMode
	// switch modeStr {
	// case "stdio":
	// 	mcpMode = openapimcp.StdIO
	// case "sse":
	// 	mcpMode = openapimcp.SSE
	// default:
	// 	log.Printf("Error: Invalid mode '%s'. Allowed modes are 'stdio' or 'sse'.", modeStr)
	// 	pflag.Usage()
	// 	os.Exit(1)
	// }

	// log.Printf("Initializing MCP server generator with spec: %s, mode: %s", specSource, mcpMode)

	config := openapimcp.GeneratorConfig{
		SpecSource: specSource,
		ServerMode: openapimcp.StdIO,
	}

	if err := openapimcp.RunFromSpec(config); err != nil {
		log.Fatalf("Error running MCP server generator: %v", err)
	}

	log.Println("Application finished.")
}

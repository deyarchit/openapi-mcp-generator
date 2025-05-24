package openapimcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

// LoadSpec loads an OpenAPI 3 specification from a given source (URL or local file path).
// It supports both JSON and YAML formats.
func LoadSpec(source string) (*openapi3.T, error) {
	if source == "" {
		return nil, fmt.Errorf("OpenAPI spec source cannot be empty")
	}

	var specData []byte
	var err error
	var baseURI *url.URL // For resolving relative references if any

	// Check if the source is a URL
	u, urlErr := url.ParseRequestURI(source)
	if urlErr == nil && (u.Scheme == "http" || u.Scheme == "https") {
		// It's a URL
		log.Debugf("Loading OpenAPI spec from URL: %s\n", source)
		resp, httpErr := http.Get(source)
		if httpErr != nil {
			return nil, fmt.Errorf("failed to fetch spec from URL %s: %w", source, httpErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch spec from URL %s: status code %d", source, resp.StatusCode)
		}

		specData, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec data from URL %s: %w", source, err)
		}
		baseURI = u
	} else {
		// Assume it's a local file path
		log.Debugf("Loading OpenAPI spec from file: %s\n", source)
		absPath, pathErr := filepath.Abs(source)
		if pathErr != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", source, pathErr)
		}
		specData, err = os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file %s: %w", source, err)
		}
		baseURI, err = url.Parse("file://" + filepath.ToSlash(absPath))
		if err != nil {
			return nil, fmt.Errorf("failed to create base URI for local file %s: %w", absPath, err)
		}
	}

	// Use kin-openapi's loader
	// The loader can automatically detect JSON or YAML.
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false // Allow external references if any

	// kin-openapi's LoadFromData requires a URI to resolve relative references,
	// even for local files.
	doc, err := loader.LoadFromDataWithPath(specData, baseURI)
	if err != nil {
		// Try to provide more specific error if it's a syntax issue
		if strings.Contains(err.Error(), "yaml:") || strings.Contains(err.Error(), "json:") {
			return nil, fmt.Errorf("failed to parse OpenAPI spec (JSON/YAML syntax error or invalid structure): %w", err)
		}
		return nil, fmt.Errorf("failed to load/parse OpenAPI spec from %s: %w", source, err)
	}

	// Validate the loaded spec (optional, but good practice)
	err = doc.Validate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("OpenAPI spec validation failed: %w", err)
	}

	log.Debugf("Successfully loaded and validated OpenAPI spec: %s (Version: %s)\n", doc.Info.Title, doc.Info.Version)
	return doc, nil
}

package api

import _ "embed"

// OpenAPI is the service HTTP contract (source of truth for Swagger).
//
//go:embed openapi.yaml
var OpenAPI []byte

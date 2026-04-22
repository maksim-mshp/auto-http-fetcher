package autofetcher

import "embed"

// OpenAPIFS contains the generated OpenAPI files from the repository root.
//
//go:embed api/openapi.json api/openapi.yml
var OpenAPIFS embed.FS

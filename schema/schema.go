// Package schema exposes the normative LogSchema JSON Schema documents.
package schema

import (
	"embed"
	"fmt"
	"io/fs"
)

const (
	// CoreV1DefinitionsID identifies the LogSchema core v1 definitions.
	CoreV1DefinitionsID = "https://tkuchiki.github.io/logschema/schema/core/v1/definitions.schema.json"
	// HTTPV1RequestID identifies the LogSchema HTTP request v1 schema.
	HTTPV1RequestID = "https://tkuchiki.github.io/logschema/schema/http/v1/request.schema.json"
	// SQLV1QueryID identifies the LogSchema SQL query v1 schema.
	SQLV1QueryID = "https://tkuchiki.github.io/logschema/schema/sql/v1/query.schema.json"
)

// Resource associates a schema identifier with its embedded file path.
type Resource struct {
	ID   string
	Path string
}

var resources = [...]Resource{
	{ID: CoreV1DefinitionsID, Path: "core/v1/definitions.schema.json"},
	{ID: HTTPV1RequestID, Path: "http/v1/request.schema.json"},
	{ID: SQLV1QueryID, Path: "sql/v1/query.schema.json"},
}

// FS contains the catalog and the core, HTTP, and SQL JSON Schema documents.
//
//go:embed catalog.json core http sql
var FS embed.FS

// Resources returns the schema resources that consumers should register with
// their validator before compiling a schema that contains external references.
func Resources() []Resource {
	return append([]Resource(nil), resources[:]...)
}

// Read returns an embedded schema resource by its identifier. It returns an
// error matching fs.ErrNotExist when the identifier is unknown.
func Read(id string) ([]byte, error) {
	for _, resource := range resources {
		if resource.ID != id {
			continue
		}

		data, err := fs.ReadFile(FS, resource.Path)
		if err != nil {
			return nil, fmt.Errorf("logschema: read schema resource %q: %w", id, err)
		}

		return data, nil
	}

	return nil, fmt.Errorf("logschema: schema resource %q: %w", id, fs.ErrNotExist)
}

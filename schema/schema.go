// Package schema exposes the normative LogSchema JSON Schema documents.
package schema

import "embed"

// FS contains the core, HTTP, and SQL JSON Schema documents.
//
//go:embed core http sql
var FS embed.FS

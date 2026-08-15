package logschema_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	httpv1 "github.com/tkuchiki/logschema/http/v1"
	logschemaschema "github.com/tkuchiki/logschema/schema"
	sqlv1 "github.com/tkuchiki/logschema/sql/v1"
)

const (
	coreSchemaID = "https://tkuchiki.github.io/logschema/schema/core/v1/definitions.schema.json"
	httpSchemaID = "https://tkuchiki.github.io/logschema/schema/http/v1/request.schema.json"
	sqlSchemaID  = "https://tkuchiki.github.io/logschema/schema/sql/v1/query.schema.json"
)

func TestConformanceFixtures(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	for id, path := range map[string]string{
		coreSchemaID: "core/v1/definitions.schema.json",
		httpSchemaID: "http/v1/request.schema.json",
		sqlSchemaID:  "sql/v1/query.schema.json",
	} {
		data, err := fs.ReadFile(logschemaschema.FS, path)
		if err != nil {
			t.Fatalf("read embedded schema %s: %v", path, err)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("decode schema %s: %v", id, err)
		}
		if err := compiler.AddResource(id, document); err != nil {
			t.Fatalf("add schema %s: %v", id, err)
		}
	}

	for _, tc := range []struct {
		name     string
		schemaID string
		root     string
		newGo    func() any
	}{
		{
			name:     "http",
			schemaID: httpSchemaID,
			root:     "testdata/http/v1",
			newGo:    func() any { return new(httpv1.Request) },
		},
		{
			name:     "sql",
			schemaID: sqlSchemaID,
			root:     "testdata/sql/v1",
			newGo:    func() any { return new(sqlv1.Query) },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := compiler.Compile(tc.schemaID)
			if err != nil {
				t.Fatalf("compile schema: %v", err)
			}
			testFixtureDirectory(t, schema, filepath.Join(tc.root, "valid"), true, tc.newGo)
			testFixtureDirectory(t, schema, filepath.Join(tc.root, "invalid"), false, tc.newGo)
		})
	}
}

func testFixtureDirectory(
	t *testing.T,
	schema *jsonschema.Schema,
	root string,
	wantValid bool,
	newGo func() any,
) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read fixtures %s: %v", root, err)
	}
	if len(entries) == 0 {
		t.Fatalf("no fixtures in %s", root)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}

			instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decode JSON: %v", err)
			}

			err = schema.Validate(instance)
			if wantValid && err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
			if !wantValid && err == nil {
				t.Fatal("invalid fixture accepted")
			}

			if !wantValid {
				record := newGo()
				if err := json.Unmarshal(data, record); err == nil {
					t.Fatal("invalid fixture accepted by Go reference decoder")
				}

				return
			}

			record := newGo()
			if err := json.Unmarshal(data, record); err != nil {
				t.Fatalf("decode Go reference type: %v", err)
			}

			roundTrip, err := json.Marshal(record)
			if err != nil {
				t.Fatalf("marshal Go reference type: %v", err)
			}

			roundTripInstance, err := jsonschema.UnmarshalJSON(bytes.NewReader(roundTrip))
			if err != nil {
				t.Fatalf("decode round-trip JSON: %v", err)
			}

			if err := schema.Validate(roundTripInstance); err != nil {
				t.Fatalf("Go reference output violates schema: %v", err)
			}
		})
	}
}

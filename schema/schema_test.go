package schema

import (
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

func TestResourcesMatchCatalogAndSchemaIDs(t *testing.T) {
	catalogData, err := fs.ReadFile(FS, "catalog.json")
	if err != nil {
		t.Fatal(err)
	}

	var catalog map[string]string
	if err := json.Unmarshal(catalogData, &catalog); err != nil {
		t.Fatal(err)
	}

	wantCatalog := make(map[string]string, len(resources))
	registeredPaths := make(map[string]bool, len(resources))
	for _, resource := range Resources() {
		if _, exists := wantCatalog[resource.ID]; exists {
			t.Fatalf("duplicate schema ID %q", resource.ID)
		}
		if registeredPaths[resource.Path] {
			t.Fatalf("duplicate schema path %q", resource.Path)
		}

		wantCatalog[resource.ID] = resource.Path
		registeredPaths[resource.Path] = true

		data, err := Read(resource.ID)
		if err != nil {
			t.Fatal(err)
		}

		var metadata struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			t.Fatal(err)
		}
		if metadata.ID != resource.ID {
			t.Errorf("schema %s has $id %q", resource.Path, metadata.ID)
		}
	}

	if !reflect.DeepEqual(catalog, wantCatalog) {
		t.Fatalf("catalog = %#v, want %#v", catalog, wantCatalog)
	}

	embeddedPaths := make(map[string]bool, len(resources))
	err = fs.WalkDir(FS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".schema.json") {
			embeddedPaths[path] = true
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(embeddedPaths, registeredPaths) {
		t.Fatalf("embedded schema paths = %#v, registered paths = %#v", embeddedPaths, registeredPaths)
	}
}

func TestResourcesReturnsCopy(t *testing.T) {
	got := Resources()
	got[0].ID = "modified"

	if Resources()[0].ID == "modified" {
		t.Fatal("Resources returned mutable registry storage")
	}
}

func TestReadRejectsUnknownID(t *testing.T) {
	_, err := Read("https://example.com/unknown.schema.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Read error = %v, want fs.ErrNotExist", err)
	}
}

package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// UnmarshalStrict decodes one JSON value, rejects unknown fields, and verifies
// that the named top-level fields are present and non-null.
func UnmarshalStrict(data []byte, destination any, requiredNonNullFields ...string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("logschema: multiple JSON values are not allowed")
		}
		return fmt.Errorf("logschema: decode trailing JSON: %w", err)
	}

	if len(requiredNonNullFields) == 0 {
		return nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	for _, name := range requiredNonNullFields {
		value, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("logschema: required field %q must be present and non-null", name)
		}
	}

	return nil
}

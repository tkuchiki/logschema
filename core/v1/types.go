package corev1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

type SourceKind string

const (
	SourceFile     SourceKind = "file"
	SourceStdin    SourceKind = "stdin"
	SourceOTLPFile SourceKind = "otlp_file"
	SourceOther    SourceKind = "other"
)

type Source struct {
	Kind        SourceKind   `json:"kind"`
	Name        *string      `json:"name"`
	Fingerprint *Fingerprint `json:"fingerprint"`

	OffsetStart *DecimalUint64 `json:"offset_start"`
	OffsetEnd   *DecimalUint64 `json:"offset_end"`
	LineStart   *DecimalUint64 `json:"line_start"`
	LineEnd     *DecimalUint64 `json:"line_end"`
}

func (s Source) Validate() error {
	switch s.Kind {
	case SourceFile, SourceStdin, SourceOTLPFile, SourceOther:
	default:
		return fmt.Errorf("logschema: unsupported source kind %q", s.Kind)
	}

	if s.Name != nil && *s.Name == "" {
		return fmt.Errorf("logschema: source name must not be empty")
	}
	if s.Fingerprint != nil {
		if err := s.Fingerprint.Validate(); err != nil {
			return err
		}
	}

	if s.LineStart != nil && *s.LineStart == 0 {
		return fmt.Errorf("logschema: source line_start must be greater than zero")
	}
	if s.LineEnd != nil && *s.LineEnd == 0 {
		return fmt.Errorf("logschema: source line_end must be greater than zero")
	}

	return nil
}

// Fingerprint identifies a value together with the versioned algorithm that
// produced it. The producing algorithm defines whether the value is safe to
// disclose.
type Fingerprint struct {
	Value     string `json:"value"`
	Algorithm string `json:"algorithm"`
	Version   string `json:"version"`
}

// Validate verifies that all fingerprint identity components are present.
func (f Fingerprint) Validate() error {
	if f.Value == "" {
		return fmt.Errorf("logschema: fingerprint value must not be empty")
	}
	if f.Algorithm == "" {
		return fmt.Errorf("logschema: fingerprint algorithm must not be empty")
	}
	if f.Version == "" {
		return fmt.Errorf("logschema: fingerprint version must not be empty")
	}

	return nil
}

type Resource struct {
	ServiceName *string    `json:"service_name"`
	Attributes  Attributes `json:"attributes"`
}

func (r Resource) Validate() error {
	if r.ServiceName != nil && *r.ServiceName == "" {
		return fmt.Errorf("logschema: resource service_name must not be empty")
	}

	return r.Attributes.Validate()
}

type TraceContext struct {
	TraceID    string  `json:"trace_id"`
	SpanID     *string `json:"span_id"`
	TraceFlags *uint8  `json:"trace_flags"`
}

func (c TraceContext) Validate() error {
	if !isValidLowerHexID(c.TraceID, 32) {
		return fmt.Errorf("logschema: trace_id must be a non-zero lowercase 32-character hexadecimal value")
	}
	if c.SpanID != nil && !isValidLowerHexID(*c.SpanID, 16) {
		return fmt.Errorf("logschema: span_id must be a non-zero lowercase 16-character hexadecimal value")
	}

	return nil
}

func isValidLowerHexID(value string, length int) bool {
	if len(value) != length {
		return false
	}

	nonZero := false
	for i := 0; i < len(value); i++ {
		if (value[i] < '0' || value[i] > '9') && (value[i] < 'a' || value[i] > 'f') {
			return false
		}
		if value[i] != '0' {
			nonZero = true
		}
	}

	return nonZero
}

type Attributes map[string]any

func (a Attributes) Validate() error {
	for key, value := range a {
		if key == "" {
			return fmt.Errorf("logschema: attribute key must not be empty")
		}
		if err := validateAttributeValue(value); err != nil {
			return fmt.Errorf("logschema: attribute %q: %w", key, err)
		}
	}

	return nil
}

// MarshalJSON validates attributes and encodes a nil map as an empty object.
func (a Attributes) MarshalJSON() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}

	if a == nil {
		return []byte("{}"), nil
	}

	type plainAttributes Attributes

	return json.Marshal(plainAttributes(a))
}

// UnmarshalJSON decodes an attribute object while preserving JSON numbers exactly.
func (a *Attributes) UnmarshalJSON(data []byte) error {
	if a == nil {
		return fmt.Errorf("logschema: Attributes receiver is nil")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var decoded map[string]any
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("logschema: decode attributes: %w", err)
	}
	if decoded == nil {
		return fmt.Errorf("logschema: attributes must be an object")
	}

	*a = Attributes(decoded)

	return nil
}

func validateAttributeValue(value any) error {
	if value == nil {
		return fmt.Errorf("value must not be null; omit the attribute key instead")
	}

	switch value := value.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return nil
	case float32:
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("floating-point value must be finite")
		}

		return nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("floating-point value must be finite")
		}

		return nil
	case json.Number:
		if _, err := json.Marshal(value); err != nil {
			return fmt.Errorf("invalid JSON number: %w", err)
		}

		return nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(rv.Float()) || math.IsInf(rv.Float(), 0) {
			return fmt.Errorf("floating-point value must be finite")
		}

		return nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return fmt.Errorf("byte slices are not attribute arrays; encode binary data explicitly as a string")
		}
	case reflect.Array:
	default:
		return fmt.Errorf("value must be a scalar or scalar array")
	}

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		itemValue := reflect.ValueOf(item)
		if itemValue.IsValid() {
			switch itemValue.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				return fmt.Errorf("array item %d must be a scalar", i)
			}
		}

		if err := validateAttributeValue(item); err != nil {
			return fmt.Errorf("array item %d: %w", i, err)
		}
	}

	return nil
}

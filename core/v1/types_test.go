package corev1

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestAttributesValidate(t *testing.T) {
	type label string

	valid := Attributes{
		"string": "value",
		"alias":  label("value"),
		"number": json.Number("42"),
		"array":  []any{"value", true},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid attributes rejected: %v", err)
	}
	invalid := Attributes{"nested": map[string]any{"key": "value"}}
	if err := invalid.Validate(); err == nil {
		t.Fatal("nested object was accepted")
	}

	nullValue := Attributes{"null": nil}
	if err := nullValue.Validate(); err == nil {
		t.Fatal("null attribute value was accepted")
	}

	nonFinite := Attributes{"number": math.Inf(1)}
	if err := nonFinite.Validate(); err == nil {
		t.Fatal("non-finite attribute value was accepted")
	}

	invalidNumber := Attributes{"number": json.Number("01")}
	if err := invalidNumber.Validate(); err == nil {
		t.Fatal("invalid JSON number was accepted")
	}

	binary := Attributes{"binary": []byte("value")}
	if err := binary.Validate(); err == nil {
		t.Fatal("byte slice was accepted without an explicit string encoding")
	}
}

func TestNilAttributesMarshalAsEmptyObject(t *testing.T) {
	data, err := json.Marshal(Attributes(nil))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Fatalf("nil attributes encoded as %s, want {}", data)
	}
}

func TestSourceValidateLineNumber(t *testing.T) {
	zero := DecimalUint64(0)

	source := Source{
		Kind:      SourceFile,
		LineStart: &zero,
	}
	if err := source.Validate(); err == nil {
		t.Fatal("zero source line was accepted")
	}
}

func TestSourceValidateFingerprint(t *testing.T) {
	source := Source{
		Kind: SourceFile,
		Fingerprint: &Fingerprint{
			Value:     "content-id",
			Algorithm: "sha256",
		},
	}

	if err := source.Validate(); err == nil {
		t.Fatal("source fingerprint without a version was accepted")
	}
}

func TestSourceValidateAttributes(t *testing.T) {
	valid := Source{
		Kind: SourceOther,
		Attributes: Attributes{
			"adapter.name":    "example",
			"adapter.version": "1.0.0",
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid source attributes rejected: %v", err)
	}

	invalid := Source{
		Kind:       SourceOther,
		Attributes: Attributes{"adapter.config": map[string]any{"nested": true}},
	}
	if err := invalid.Validate(); err == nil {
		t.Fatal("nested source attribute was accepted")
	}
}

func TestSourceMarshalNilAttributesAsEmptyObject(t *testing.T) {
	data, err := json.Marshal(Source{Kind: SourceStdin})
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	attributes, ok := decoded["attributes"].(map[string]any)
	if !ok || len(attributes) != 0 {
		t.Fatalf("source attributes = %#v, want empty object", decoded["attributes"])
	}
}

func TestAttributesMarshalRejectsInvalidValue(t *testing.T) {
	_, err := json.Marshal(Attributes{"nested": map[string]any{"key": "value"}})
	if err == nil {
		t.Fatal("invalid attributes were marshaled")
	}
}

func TestTraceContextValidate(t *testing.T) {
	traceState := "vendor=value,other=state"
	valid := TraceContext{
		TraceID:    "4bf92f3577b34da6a3ce929d0e0e4736",
		TraceState: &traceState,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid trace context was rejected: %v", err)
	}

	zero := TraceContext{TraceID: "00000000000000000000000000000000"}
	if err := zero.Validate(); err == nil {
		t.Fatal("zero trace ID was accepted")
	}

	uppercase := TraceContext{TraceID: "4BF92F3577B34DA6A3CE929D0E0E4736"}
	if err := uppercase.Validate(); err == nil {
		t.Fatal("uppercase trace ID was accepted")
	}
}

func TestTraceContextValidateTraceState(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValid bool
	}{
		{name: "typical", value: "vendor=value,other=state", wantValid: true},
		{name: "horizontal tab", value: "vendor=value,\tother=state", wantValid: true},
		{name: "maximum length", value: strings.Repeat("a", 512), wantValid: true},
		{name: "empty"},
		{name: "oversized", value: strings.Repeat("a", 513)},
		{name: "control character", value: "vendor=value\n"},
		{name: "non-ASCII", value: "vendor=値"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := TraceContext{
				TraceID:    "4bf92f3577b34da6a3ce929d0e0e4736",
				TraceState: &test.value,
			}
			err := context.Validate()
			if test.wantValid && err != nil {
				t.Fatalf("valid trace state was rejected: %v", err)
			}
			if !test.wantValid && err == nil {
				t.Fatal("invalid trace state was accepted")
			}
		})
	}
}

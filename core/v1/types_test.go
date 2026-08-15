package corev1

import (
	"encoding/json"
	"math"
	"testing"
)

func TestAttributesValidate(t *testing.T) {
	valid := Attributes{
		"string": "value",
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

func TestTraceContextValidate(t *testing.T) {
	valid := TraceContext{TraceID: "4bf92f3577b34da6a3ce929d0e0e4736"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("trace ID without span ID was rejected: %v", err)
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

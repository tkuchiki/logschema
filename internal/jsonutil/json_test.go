package jsonutil

import "testing"

type testRecord struct {
	Required string `json:"required"`
}

func TestUnmarshalStrict(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid",
			input: `{"required":"value"}`,
			valid: true,
		},
		{
			name:  "missing required",
			input: `{}`,
			valid: false,
		},
		{
			name:  "null required",
			input: `{"required":null}`,
			valid: false,
		},
		{
			name:  "unknown field",
			input: `{"required":"value","unknown":true}`,
			valid: false,
		},
		{
			name:  "multiple values",
			input: `{"required":"value"} {}`,
			valid: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var record testRecord
			err := UnmarshalStrict([]byte(tc.input), &record, "required")

			if tc.valid && err != nil {
				t.Fatalf("valid input rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

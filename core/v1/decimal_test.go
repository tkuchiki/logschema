package corev1

import (
	"encoding/json"
	"math"
	"testing"
)

func TestDecimalUint64RoundTrip(t *testing.T) {
	want := DecimalUint64(math.MaxUint64)
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"18446744073709551615"` {
		t.Fatalf("unexpected encoding: %s", data)
	}
	var got DecimalUint64
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func TestDecimalRejectsJSONNumber(t *testing.T) {
	for _, value := range []string{`42`, `"01"`, `"-0"`, `"+1"`} {
		var decoded DecimalUint64
		if err := json.Unmarshal([]byte(value), &decoded); err == nil {
			t.Errorf("invalid decimal %s was accepted", value)
		}
	}
}

func TestDecimalBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		new   func() any
		valid bool
	}{
		{
			name:  "maximum uint64",
			value: `"18446744073709551615"`,
			new:   func() any { return new(DecimalUint64) },
			valid: true,
		},
		{
			name:  "uint64 overflow",
			value: `"18446744073709551616"`,
			new:   func() any { return new(DecimalUint64) },
			valid: false,
		},
		{
			name:  "maximum int64",
			value: `"9223372036854775807"`,
			new:   func() any { return new(DecimalInt64) },
			valid: true,
		},
		{
			name:  "minimum int64",
			value: `"-9223372036854775808"`,
			new:   func() any { return new(DecimalInt64) },
			valid: true,
		},
		{
			name:  "int64 overflow",
			value: `"9223372036854775808"`,
			new:   func() any { return new(DecimalInt64) },
			valid: false,
		},
		{
			name:  "int64 underflow",
			value: `"-9223372036854775809"`,
			new:   func() any { return new(DecimalInt64) },
			valid: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tc.value), tc.new())
			if tc.valid && err != nil {
				t.Fatalf("valid boundary rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("invalid boundary accepted")
			}
		})
	}
}

func TestIsCanonicalDecimal(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{value: `"0"`, want: true},
		{value: `"42"`, want: true},
		{value: `"-42"`, want: true},
		{value: `"01"`, want: false},
		{value: `"-0"`, want: false},
		{value: `42`, want: false},
	} {
		if got := IsCanonicalDecimal([]byte(tc.value)); got != tc.want {
			t.Errorf("IsCanonicalDecimal(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

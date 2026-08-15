// Package corev1 provides storage-neutral common reference types for LogSchema v1.
package corev1

import (
	"fmt"
	"strconv"
)

// DecimalInt64 is encoded as a quoted base-10 integer in JSON.
type DecimalInt64 int64

func (v DecimalInt64) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, strconv.FormatInt(int64(v), 10)), nil
}

func (v *DecimalInt64) UnmarshalJSON(data []byte) error {
	if v == nil {
		return fmt.Errorf("logschema: DecimalInt64 receiver is nil")
	}

	raw, err := unmarshalCanonicalDecimal(data, "signed")
	if err != nil {
		return err
	}

	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("logschema: invalid signed decimal %q: %w", raw, err)
	}

	*v = DecimalInt64(n)

	return nil
}

// DecimalUint64 is encoded as a quoted base-10 unsigned integer in JSON.
type DecimalUint64 uint64

func (v DecimalUint64) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, strconv.FormatUint(uint64(v), 10)), nil
}

func (v *DecimalUint64) UnmarshalJSON(data []byte) error {
	if v == nil {
		return fmt.Errorf("logschema: DecimalUint64 receiver is nil")
	}

	raw, err := unmarshalCanonicalDecimal(data, "unsigned")
	if err != nil {
		return err
	}

	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("logschema: invalid unsigned decimal %q: %w", raw, err)
	}

	*v = DecimalUint64(n)

	return nil
}

func unmarshalCanonicalDecimal(data []byte, kind string) (string, error) {
	if !IsCanonicalDecimal(data) {
		return "", fmt.Errorf("logschema: %s 64-bit values must be canonical decimal JSON strings", kind)
	}

	return string(data[1 : len(data)-1]), nil
}

// IsCanonicalDecimal reports whether data is a canonical quoted decimal.
func IsCanonicalDecimal(data []byte) bool {
	if len(data) < 3 || data[0] != '"' || data[len(data)-1] != '"' {
		return false
	}

	raw := data[1 : len(data)-1]
	if len(raw) == 1 && raw[0] == '0' {
		return true
	}

	if raw[0] == '-' {
		raw = raw[1:]
		if len(raw) == 1 && raw[0] == '0' {
			return false
		}
	}

	if len(raw) == 0 || raw[0] == '0' {
		return false
	}

	for _, b := range raw {
		if b < '0' || b > '9' {
			return false
		}
	}

	return true
}

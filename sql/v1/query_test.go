package sqlv1

import (
	"testing"

	corev1 "github.com/tkuchiki/logschema/core/v1"
)

func TestNewQuery(t *testing.T) {
	record, err := NewQuery(
		corev1.DecimalUint64(850_000_000),
		corev1.Source{Kind: corev1.SourceStdin},
		QueryData{
			DBSystem: "postgresql",
			Fingerprint: Fingerprint{
				Value:     "update-jobs",
				Algorithm: "slp",
				Version:   "1",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.SchemaVersion != SchemaVersion || record.Kind != Kind {
		t.Fatalf("constructor metadata = %q, %q", record.SchemaVersion, record.Kind)
	}
}

func TestFingerprintValidate(t *testing.T) {
	fingerprint := Fingerprint{Value: "select-one", Algorithm: "slp"}
	if err := fingerprint.Validate(); err == nil {
		t.Fatal("fingerprint without version was accepted")
	}
}

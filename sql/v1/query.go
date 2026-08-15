// Package sqlv1 provides the normalized database slow-query log record v1.
package sqlv1

import (
	"encoding/json"
	"fmt"

	corev1 "github.com/tkuchiki/logschema/core/v1"
	"github.com/tkuchiki/logschema/internal/jsonutil"
)

const (
	SchemaVersion = "logschema.sql.query/v1"
	Kind          = "sql.query"
)

type Query struct {
	SchemaVersion        string               `json:"schema_version"`
	Kind                 string               `json:"kind"`
	TimeUnixNano         *corev1.DecimalInt64 `json:"time_unix_nano"`
	ObservedTimeUnixNano *corev1.DecimalInt64 `json:"observed_time_unix_nano"`
	DurationNano         corev1.DecimalUint64 `json:"duration_nano"`
	Source               corev1.Source        `json:"source"`
	Resource             corev1.Resource      `json:"resource"`
	TraceContext         *corev1.TraceContext `json:"trace_context"`
	Data                 QueryData            `json:"data"`
}

type QueryData struct {
	DBSystem         string                `json:"db_system"`
	DatabaseName     *string               `json:"database_name"`
	Operation        *string               `json:"operation"`
	QueryText        *string               `json:"query_text"`
	QuerySummary     *string               `json:"query_summary"`
	Fingerprint      Fingerprint           `json:"fingerprint"`
	LockDurationNano *corev1.DecimalUint64 `json:"lock_duration_nano"`
	RowsSent         *corev1.DecimalUint64 `json:"rows_sent"`
	RowsExamined     *corev1.DecimalUint64 `json:"rows_examined"`
	RowsAffected     *corev1.DecimalUint64 `json:"rows_affected"`
	BytesSent        *corev1.DecimalUint64 `json:"bytes_sent"`
	Attributes       corev1.Attributes     `json:"attributes"`
}

// Fingerprint is the shared versioned fingerprint representation.
type Fingerprint = corev1.Fingerprint

// NewQuery creates and validates a SQL query record with canonical metadata.
func NewQuery(
	durationNano corev1.DecimalUint64,
	source corev1.Source,
	data QueryData,
) (Query, error) {
	record := Query{
		SchemaVersion: SchemaVersion,
		Kind:          Kind,
		DurationNano:  durationNano,
		Source:        source,
		Data:          data,
	}
	if err := record.Validate(); err != nil {
		return Query{}, err
	}

	return record, nil
}

func (q *Query) UnmarshalJSON(data []byte) error {
	type plainQuery Query

	var decoded plainQuery
	if err := jsonutil.UnmarshalStrict(data, &decoded, "duration_nano", "resource"); err != nil {
		return err
	}

	record := Query(decoded)
	if err := record.Validate(); err != nil {
		return err
	}

	*q = record

	return nil
}

// MarshalJSON validates and encodes a SQL query record.
func (q Query) MarshalJSON() ([]byte, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	type plainQuery Query

	return json.Marshal(plainQuery(q))
}

func (q Query) Validate() error {
	if q.SchemaVersion != SchemaVersion {
		return fmt.Errorf("logschema: unsupported SQL query schema %q", q.SchemaVersion)
	}
	if q.Kind != Kind {
		return fmt.Errorf("logschema: unsupported SQL query kind %q", q.Kind)
	}

	if err := q.Source.Validate(); err != nil {
		return err
	}
	if err := q.Resource.Validate(); err != nil {
		return err
	}
	if q.TraceContext != nil {
		if err := q.TraceContext.Validate(); err != nil {
			return err
		}
	}

	if q.Data.DBSystem == "" {
		return fmt.Errorf("logschema: db_system must not be empty")
	}

	if q.Data.DatabaseName != nil && *q.Data.DatabaseName == "" {
		return fmt.Errorf("logschema: database_name must not be empty")
	}
	if q.Data.Operation != nil && *q.Data.Operation == "" {
		return fmt.Errorf("logschema: SQL operation must not be empty")
	}

	if q.Data.QueryText != nil && *q.Data.QueryText == "" {
		return fmt.Errorf("logschema: query_text must not be empty")
	}
	if q.Data.QuerySummary != nil && *q.Data.QuerySummary == "" {
		return fmt.Errorf("logschema: query_summary must not be empty")
	}
	if err := q.Data.Fingerprint.Validate(); err != nil {
		return err
	}

	return q.Data.Attributes.Validate()
}

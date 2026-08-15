// Package httpv1 provides the normalized HTTP request log record v1.
package httpv1

import (
	"encoding/json"
	"fmt"
	"strings"

	corev1 "github.com/tkuchiki/logschema/core/v1"
	"github.com/tkuchiki/logschema/internal/jsonutil"
)

const (
	SchemaVersion = "logschema.http.request/v1"
	Kind          = "http.request"
)

type Request struct {
	SchemaVersion        string               `json:"schema_version"`
	Kind                 string               `json:"kind"`
	TimeUnixNano         *corev1.DecimalInt64 `json:"time_unix_nano"`
	ObservedTimeUnixNano *corev1.DecimalInt64 `json:"observed_time_unix_nano"`
	DurationNano         corev1.DecimalUint64 `json:"duration_nano"`
	Source               corev1.Source        `json:"source"`
	Resource             corev1.Resource      `json:"resource"`
	TraceContext         *corev1.TraceContext `json:"trace_context"`
	Data                 RequestData          `json:"data"`
}

type RequestData struct {
	Method                string                `json:"method"`
	URLScheme             *string               `json:"url_scheme"`
	URLAuthority          *string               `json:"url_authority"`
	URLPath               string                `json:"url_path"`
	URLQuery              *string               `json:"url_query"`
	Route                 *string               `json:"route"`
	StatusCode            *int                  `json:"status_code"`
	RequestBodySizeBytes  *corev1.DecimalUint64 `json:"request_body_size_bytes"`
	ResponseBodySizeBytes *corev1.DecimalUint64 `json:"response_body_size_bytes"`
	Attributes            corev1.Attributes     `json:"attributes"`
}

// NewRequest creates and validates an HTTP request record with canonical metadata.
func NewRequest(
	durationNano corev1.DecimalUint64,
	source corev1.Source,
	data RequestData,
) (Request, error) {
	record := Request{
		SchemaVersion: SchemaVersion,
		Kind:          Kind,
		DurationNano:  durationNano,
		Source:        source,
		Data:          data,
	}
	if err := record.Validate(); err != nil {
		return Request{}, err
	}

	return record, nil
}

// URI returns the request target derived from the path and query components.
func (d RequestData) URI() string {
	if d.URLQuery == nil {
		return d.URLPath
	}

	return d.URLPath + "?" + *d.URLQuery
}

// URLFull returns an absolute URL when both scheme and authority are available.
func (d RequestData) URLFull() (string, bool) {
	if d.URLScheme == nil || d.URLAuthority == nil || *d.URLScheme == "" || *d.URLAuthority == "" {
		return "", false
	}
	if !strings.HasPrefix(d.URLPath, "/") {
		return "", false
	}

	return *d.URLScheme + "://" + *d.URLAuthority + d.URI(), true
}

// URLReference returns the most complete URL reference represented by the
// normalized components.
func (d RequestData) URLReference() string {
	if fullURL, ok := d.URLFull(); ok {
		return fullURL
	}
	if d.URLAuthority != nil {
		return "//" + *d.URLAuthority + d.URI()
	}

	return d.URI()
}

func (r *Request) UnmarshalJSON(data []byte) error {
	type plainRequest Request

	var decoded plainRequest
	if err := jsonutil.UnmarshalStrict(data, &decoded, "duration_nano", "resource"); err != nil {
		return err
	}

	record := Request(decoded)
	if err := record.Validate(); err != nil {
		return err
	}

	*r = record

	return nil
}

// MarshalJSON validates and encodes an HTTP request record.
func (r Request) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	type plainRequest Request

	return json.Marshal(plainRequest(r))
}

func (r Request) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("logschema: unsupported HTTP request schema %q", r.SchemaVersion)
	}
	if r.Kind != Kind {
		return fmt.Errorf("logschema: unsupported HTTP request kind %q", r.Kind)
	}

	if err := r.Source.Validate(); err != nil {
		return err
	}
	if err := r.Resource.Validate(); err != nil {
		return err
	}
	if r.TraceContext != nil {
		if err := r.TraceContext.Validate(); err != nil {
			return err
		}
	}

	if !isHTTPToken(r.Data.Method) {
		return fmt.Errorf("logschema: HTTP method must be a non-empty HTTP token")
	}
	if r.Data.URLPath == "" {
		return fmt.Errorf("logschema: HTTP url_path must not be empty")
	}
	if r.Data.URLScheme != nil && !isValidURLScheme(*r.Data.URLScheme) {
		return fmt.Errorf("logschema: HTTP url_scheme must be a valid URI scheme")
	}
	if r.Data.URLAuthority != nil && !isSafeURLAuthority(*r.Data.URLAuthority) {
		return fmt.Errorf("logschema: HTTP url_authority must be non-empty and must not contain user information, URL component delimiters, or control characters")
	}

	if r.Data.Route != nil && *r.Data.Route == "" {
		return fmt.Errorf("logschema: HTTP route must not be empty")
	}
	if r.Data.StatusCode != nil && (*r.Data.StatusCode < 100 || *r.Data.StatusCode > 599) {
		return fmt.Errorf("logschema: HTTP status_code must be between 100 and 599")
	}

	return r.Data.Attributes.Validate()
}

func isHTTPToken(value string) bool {
	if value == "" {
		return false
	}

	for i := 0; i < len(value); i++ {
		char := value[i]
		if isASCIIAlpha(char) || (char >= '0' && char <= '9') {
			continue
		}

		switch char {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}

	return true
}

func isSafeURLAuthority(authority string) bool {
	if authority == "" {
		return false
	}

	for _, char := range authority {
		if char <= ' ' || (char >= 0x7f && char <= 0x9f) {
			return false
		}

		switch char {
		case '@', '/', '?', '#':
			return false
		}
	}

	return true
}

func isValidURLScheme(scheme string) bool {
	if len(scheme) == 0 || !isASCIIAlpha(scheme[0]) {
		return false
	}

	for i := 1; i < len(scheme); i++ {
		char := scheme[i]
		if !isASCIIAlpha(char) && (char < '0' || char > '9') && char != '+' && char != '-' && char != '.' {
			return false
		}
	}

	return true
}

func isASCIIAlpha(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

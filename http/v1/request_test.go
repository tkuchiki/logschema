package httpv1

import (
	"encoding/json"
	"testing"

	corev1 "github.com/tkuchiki/logschema/core/v1"
)

func TestRequestDataURI(t *testing.T) {
	empty := ""
	query := "q=logschema&page=2"

	tests := []struct {
		name  string
		query *string
		want  string
	}{
		{name: "without query", want: "/search"},
		{name: "empty query", query: &empty, want: "/search?"},
		{name: "with query", query: &query, want: "/search?q=logschema&page=2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := RequestData{URLPath: "/search", URLQuery: test.query}
			if got := data.URI(); got != test.want {
				t.Fatalf("URI() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRequestDataURLFull(t *testing.T) {
	scheme := "https"
	authority := "example.com:8443"
	query := "view=compact"

	data := RequestData{
		URLScheme:    &scheme,
		URLAuthority: &authority,
		URLPath:      "/users/42",
		URLQuery:     &query,
	}
	got, ok := data.URLFull()
	if !ok {
		t.Fatal("URLFull() reported that an absolute URL was unavailable")
	}
	if want := "https://example.com:8443/users/42?view=compact"; got != want {
		t.Fatalf("URLFull() = %q, want %q", got, want)
	}

	data.URLAuthority = nil
	if _, ok := data.URLFull(); ok {
		t.Fatal("URLFull() succeeded without an authority")
	}

	data.URLAuthority = &authority
	data.URLPath = "*"
	if _, ok := data.URLFull(); ok {
		t.Fatal("URLFull() succeeded for a non-path request target")
	}
}

func TestRequestDataURLReference(t *testing.T) {
	scheme := "https"
	authority := "example.com"
	query := "page=2"

	tests := []struct {
		name string
		data RequestData
		want string
	}{
		{
			name: "origin form",
			data: RequestData{URLPath: "/search", URLQuery: &query},
			want: "/search?page=2",
		},
		{
			name: "network path reference",
			data: RequestData{URLAuthority: &authority, URLPath: "/search"},
			want: "//example.com/search",
		},
		{
			name: "absolute URL",
			data: RequestData{URLScheme: &scheme, URLAuthority: &authority, URLPath: "/search"},
			want: "https://example.com/search",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.data.URLReference(); got != test.want {
				t.Fatalf("URLReference() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	record, err := NewRequest(
		corev1.DecimalUint64(125_000_000),
		corev1.Source{Kind: corev1.SourceStdin},
		RequestData{Method: "GET", URLPath: "/healthz"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.SchemaVersion != SchemaVersion || record.Kind != Kind {
		t.Fatalf("constructor metadata = %q, %q", record.SchemaVersion, record.Kind)
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) {
		t.Fatalf("constructor output is invalid JSON: %s", encoded)
	}
}

func TestRequestValidateHTTPMethodToken(t *testing.T) {
	record, err := NewRequest(
		1,
		corev1.Source{Kind: corev1.SourceStdin},
		RequestData{Method: "M-SEARCH", URLPath: "*"},
	)
	if err != nil {
		t.Fatalf("valid extension method was rejected: %v", err)
	}

	record.Data.Method = "GET /"
	if err := record.Validate(); err == nil {
		t.Fatal("method containing a separator was accepted")
	}
}

func TestRequestValidateURLAuthority(t *testing.T) {
	authority := "example.com/path"
	_, err := NewRequest(
		1,
		corev1.Source{Kind: corev1.SourceStdin},
		RequestData{Method: "GET", URLAuthority: &authority, URLPath: "/"},
	)
	if err == nil {
		t.Fatal("authority containing a path delimiter was accepted")
	}
}

func TestRequestMarshalRejectsInvalidRecord(t *testing.T) {
	_, err := json.Marshal(Request{})
	if err == nil {
		t.Fatal("invalid request was marshaled")
	}
}

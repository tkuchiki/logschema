# LogSchema

LogSchema provides versioned, storage-neutral schemas for structured log data.
It is intended to be shared by log producers, profilers, local query tools, and
other consumers without coupling them to a parser or storage engine.

The initial scope is deliberately small:

- common source, resource, attribute, and trace-correlation types;
- normalized HTTP request log records;
- normalized database slow-query log records;
- JSON Schema Draft 2020-12 definitions;
- pure Go reference types;
- valid and invalid conformance fixtures.

LogSchema does not provide parsers, SQL normalization, aggregation algorithms,
storage, query execution, networking, or a CLI.

## Status

LogSchema is pre-v1. Schemas and Go APIs may change incompatibly while ALP,
SLP, and the planned local workspace integration are being validated.

## Repository layout

    core/v1/       Pure Go common reference types
    http/v1/       Pure Go HTTP request reference types
    sql/v1/        Pure Go database slow-query reference types
    schema/        Normative JSON Schemas and an embedded fs.FS
    testdata/      Cross-implementation conformance fixtures

The JSON Schemas are normative. Go packages are reference implementations and
must pass the same conformance fixtures.

The Go record types use strict JSON decoding: unknown fields, non-canonical
decimal values, missing required values, and semantically invalid records are
rejected. Encoding validates records built or modified through direct struct
access before producing JSON.

The Go packages also provide small validating constructors for new records.
They set `schema_version` and `kind`; direct struct construction remains
available when a producer needs to populate optional fields before validation.

## Resolving schema references offline

Schema `$id` values are stable identifiers. Validation does not require those
HTTPS identifiers to be fetched over the network. Consumers should register all
embedded resources with their validator before compiling an HTTP or SQL schema.

Go consumers can discover and read resources without hard-coding identifiers or
repository paths:

```go
compiler := jsonschema.NewCompiler()
for _, resource := range schema.Resources() {
    data, err := schema.Read(resource.ID)
    if err != nil {
        return err
    }

    document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
    if err != nil {
        return err
    }
    if err := compiler.AddResource(resource.ID, document); err != nil {
        return err
    }
}

requestSchema, err := compiler.Compile(schema.HTTPV1RequestID)
```

Non-Go consumers can use [`schema/catalog.json`](./schema/catalog.json) to map
the same identifiers to files distributed in this repository. Validators must
use the catalog or an equivalent local registry; network retrieval is neither
required nor assumed.

## Encoding rules

- Nanosecond timestamps and 64-bit measurements are decimal strings in JSON,
  avoiding loss of precision in JSON consumers.
- `time_unix_nano` is the event time. `observed_time_unix_nano` is the time at
  which a producer observed or collected the record. Producers must not place
  an event timestamp in the observed-time field.
- Units are part of field semantics and names; durations use nanoseconds and
  sizes use bytes.
- For nullable fixed fields, an omitted property and an explicit null both mean
  that no value is available. Zero remains a real value.
- Reference encoders emit known nullable fields as null, while decoders accept
  those fields when omitted.
- Nil Go attribute maps are valid and encode as `{}`. Attribute properties may
  be omitted from input records when they are empty.
- Attribute values cannot be null. Omit the attribute key when no value is
  available.
- Go producers must encode binary attribute values explicitly as strings;
  `[]byte` values are rejected to avoid implicit base64 conversion.
- Source fingerprints include value, algorithm, and version so consumers do not
  compare identities produced by incompatible algorithms.
- HTTP URLs are stored as scheme, authority, path, and query components. The Go
  `URI`, `URLFull`, and `URLReference` methods derive combined forms so duplicate
  serialized representations cannot disagree. Producers should redact sensitive
  query values before populating `url_query`.
- SQL `query_summary` is the normalized, low-cardinality grouping form.
  `query_text` is optional because it can contain credentials, personal data,
  or other literals; producers should retain it only under an explicit policy.
- SQL fingerprints include value, algorithm, and version as one object so the
  grouping identity remains self-describing. Fingerprint values are not assumed
  to be safe: an identity algorithm can retain the complete query text.
- Unknown fields are rejected within a schema version.

## ALP and SLP integration

LogSchema is designed to be the canonical parsed-record representation in ALP
and SLP. Parsers, filters, and aggregators consume the LogSchema records
directly; consumer-specific values such as floating-point seconds are derived
only at query and rendering boundaries:

- ALP's URI grouping key uses `URLReference()`; its response time and body size
  come from `duration_nano` and `response_body_size_bytes`.
- SLP's default abstract query maps to `query_summary`. `--noabstract` input
  maps to the sensitive `query_text` field. Existing row and byte metrics map
  directly, and unavailable metrics remain null rather than becoming measured
  zeros.

Seconds-to-nanoseconds and floating-point byte conversions are producer
responsibilities. ALP and SLP integrations must define and test their rounding,
overflow, and invalid-input behavior before replacing their internal types.

See [COMPATIBILITY.md](./COMPATIBILITY.md) for versioning rules.

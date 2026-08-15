# Compatibility policy

LogSchema is currently pre-v1. This document describes the intended policy and
will become a compatibility guarantee at v1.0.0.

## Schema identity

Every record contains a required `schema_version`, for example
`logschema.http.request/v1`. Schema identity is independent from the LogSchema
Go module release version.

Schema `$id` values are stable identifiers, not required network locations.
Consumers resolve references from the bundled registry or `schema/catalog.json`
and must not require network access for validation.

## Changes within a published schema version

The JSON decoders reject unknown fields, so adding even an optional property can
break an existing consumer. Starting with the first tagged release, a published
schema identity such as `logschema.http.request/v1` is therefore frozen except
for non-semantic corrections such as descriptions and examples.

The LogSchema Go module remains pre-v1 and its Go API does not receive a v1
compatibility guarantee until the module reaches v1.0.0. Schema identities and
Go module versions are intentionally independent.

Non-semantic changes within an existing schema version may:

- improve descriptions and examples without changing semantics;
- add conformance fixtures that exercise already-defined behavior;
- fix tooling or documentation that does not change accepted records.

## Incompatible changes

The following require a new schema version:

- changing a field's meaning, type, unit, or nullability;
- removing or renaming a field;
- making an optional field required;
- changing identifier or normalization semantics;
- adding a property, including an optional property;
- adding an enum value;
- changing the interpretation of aggregate values.

## Consumer behavior

Consumers must report unsupported schema versions explicitly. They must not
silently reinterpret a record using a different version.

For nullable fixed fields, omission and an explicit null have the same meaning.
Required fields identify the minimum information needed for a record to be
interpretable. Attribute values are never null; producers omit an attribute key
when its value is unavailable.

Combined URL values are derived, not serialized. Consumers use `URI()` for the
path-plus-query form, `URLFull()` when scheme and authority are both present,
and `URLReference()` when they need the most complete available representation.
SQL query text is explicitly sensitive and must not be retained or exported by
default merely because the schema can represent it. Query fingerprint values
must receive the same treatment unless the producing algorithm guarantees that
they contain no sensitive query text.

Workspace artifacts should record:

- every schema version present;
- producer name and version;
- normalization and fingerprint algorithm versions;
- source fingerprints and redaction policy.

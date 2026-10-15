# Architecture

Implemented first slice: the application service owns connections, dedicated editor sessions, metadata, and cancellable operations; the database package handles lossless value conversion and identifier quoting. Vue uses a connection/explorer/tab workspace, lazy-loaded CodeMirror 6, and TanStack row/column virtualization. Split session/metadata packages as complexity grows rather than introducing empty layers. The current facade uses GetOperation polling every 180 ms instead of completion events; completed immutable results are retrieved once by the UI.

Execution mode is explicit: Rows uses QueryContext; Statement uses ExecContext for affected-row counts. Both can modify data. No SQL is automatically replayed to infer a mode. SQL tabs are connection-bound; changing the explorer schema affects new tabs only. Closed editor connections are physically discarded to prevent session/transaction state leaking into pooled metadata work.

See [build-guide.md](build-guide.md) for the system diagram, file responsibilities, execution flow, build procedure, and implementation checklist. This document defines the database and boundary constraints in more detail.

Go owns transport, credentials, sessions, execution, cancellation, and value conversion. Vue owns presentation and editor/navigation state. Wails exposes a small typed facade with generated bindings; driver objects and DSNs never cross it.

Introduce packages as needed: `internal/application` for use cases/DTOs, `internal/database` for driver/metadata/codecs, `internal/session` for lifecycle/cancellation, and later `internal/preferences` and `internal/credentials`. In `frontend/src`, use `components`, feature folders for connections/browser/editor/results, and a typed bridge service. Prefer Vue script/template/style single-file components. Do not create empty abstractions ahead of features.

Proposed driver: database/sql + go-sql-driver/mysql, validated on both engines before adoption. No ORM. Build configuration through the driver's config API. Set connection/read/write timeouts; verify TLS and never silently downgrade it. Initial passwords stay in memory; later persist only in OS credential storage. Settings hold opaque credential references. Handle missing/locked secrets and deletion explicitly. SQL history is sensitive, opt-in, and needs retention/deletion controls.

## Session lifecycle

A server profile is distinct from an editor session. Each editor owns a dedicated sql.Conn, serializing execution to preserve session variables and future transaction state. Metadata/table inspection uses separate connections. Bound sessions and concurrency. Future transactions belong to that same editor connection, with explicit begin/commit/rollback state; DDL can implicitly commit.

Proposed facade: Connect, Disconnect, ListSchemas, ListTables, InspectTable, StartExecution, GetExecutionResult, CancelExecution. Register an opaque operation ID before starting work; return it promptly. Events carry IDs/status, with result retrieval available after a missed event. Validate handle ownership and invalidate stale responses. Use contexts and close rows/connections/cancel functions on all paths and shutdown.

Cancellation can close the driver connection. Mark lost sessions unusable, report uncertain write/transaction outcomes honestly, and require explicit reconnect. Never retry user SQL automatically. Timeout/cancel is not proof a write did not happen. Errors use stable codes and safe messages without credentials, DSNs, or raw SQL in diagnostics.

## Lossless boundary

Table inspections additionally return optional TableMetadata with engine, primary-key membership, and up to 256 columns describing data type, nullability, and generated status. Arbitrary SQL results do not carry table eligibility. Structural candidacy currently requires an InnoDB base table and primary key; all previews remain read-only. Future apply must independently revalidate identity, schema, privileges, supported values, and original row state. The final serialized inspection result is checked against the 5 MiB cap after metadata is attached.

Use tagged cell DTOs: null, integer, decimal, float, text, temporal, json, binary, unsupported. Non-null payloads are strings; binary uses base64. Retain original column metadata and integer signedness. Use positional column IDs so duplicate column names work. Never convert BIGINT or DECIMAL to JavaScript Number. Keep temporal server text including zero dates/fractional seconds and raw JSON text; keep original payload separately from formatted display. Unknown types retain bytes and metadata instead of guessing.

Validate scanning with parseTime=false: parseTime=true converts dates into time.Time, including zero-date normalization. Test BIT, unsigned BIGINT, decimal scale, invalid UTF-8, timezone/session behavior, JSON differences, and binary against both engines. Clipped/unavailable values must be explicitly marked and cannot be copied as complete originals.

## Bounds

Proposed defaults requiring fixture validation: 500 rows, 5 MiB serialized result, 256 KiB per cell, 30-second execution timeout. Enforce in Go before the bridge, including base64/JSON overhead. Bound retained frontend data as well as virtualizing the DOM.

Do not append LIMIT to arbitrary SQL. Stop scanning at caps, close rows promptly, and report truncation; this does not bound server work. Keep multiStatements disabled and explicitly reject/handle unexpected extra result sets. Stored procedure/multiple-result support is deferred.

For table inspection, quote discovered identifiers with escaped backticks and bind user values. Prefer keyset pagination with primary/unique-key ordering. Without a suitable key, offer bounded inspection with an unstable-order warning; concurrent writes can change pages.

References: [driver documentation](https://github.com/go-sql-driver/mysql) covers cancellation and parseTime; [Wails v2 layout](https://v2.wails.io/docs/gettingstarted/firstproject/) describes build inputs.

# UI direction and queued editing

User direction recorded 2026-10-15: design the UI around easy navigation, clear data inspection/editing, a connection manager, SQL explorer, and SQL editor. Queued table edits applied with Ctrl+S or a visible button are a core daily-use workflow. The connection manager/explorer/editor/read-only grid are implemented for desktop testing; queued editing is still planned.

## Connection manager

Start with a compact connection list and a connection form: name, host, port, user, password, optional default schema, and discoverable TLS settings. Offer Test Connection and Connect, with progress and useful inline errors. Password persistence only appears once OS credential storage exists. A clear connection switcher in the workspace opens the manager again.

## Connected workspace

Keep one clear workspace per active connection, with its name and selected schema always visible. Preserve connection-specific tabs when switching without mixing their results or pending changes.

```text
Querylight    [Connection / schema ▾]                 New SQL
┌────────────────────┬──────────────────────────────────────┐
│ Search objects      │ users •  |  SQL 1                    │
│ Schema explorer     ├──────────────────────────────────────┤
│ ▾ database          │ Table: filter / refresh / row actions │
│   ▾ Tables          │                                      │
│     users           │ Virtualized data grid                │
│     inventory       │                                      │
│   ▸ Views           │                                      │
│                    ├──────────────────────────────────────┤
│                    │ 3 pending changes  Review  Discard    │
│                    │                    Apply Changes ^S  │
└────────────────────┴──────────────────────────────────────┘
Connected / status             fetched rows / limits / time
```

Use the approved mint Q/light identity with restrained accents, clear typography, and comfortable density. A focused row/cell should be easy to track. Support system light/dark preference when implemented; use readable contrast in both. Avoid dense permanent toolbars. Context menus expose occasional actions; common actions remain visible.

The sidebar is the SQL explorer: searchable schemas, tables, and views. Opening a table creates a data tab, with Structure discoverable within it later. SQL editor tabs place Run/Cancel above the editor and results below in a resizable pane. SQL text execution uses Ctrl+Enter; Ctrl+S never executes arbitrary editor SQL. Define editor file-save behavior separately when saved SQL files exist.

## Queued data edits

Editing a table cell changes a local draft. No write happens until Apply Changes or Ctrl+S in the active table tab. The apply label avoids confusing draft submission with the SQL COMMIT command. Pending changes belong to their table tab, not to a connection-wide invisible queue.

- Highlight changed cells and label added/deleted rows; show the pending row-operation count on the tab and action bar.
- Edit NULL explicitly and distinguish it from empty text. Use an expanded editor for multiline/JSON values. Binary and unsupported/clipped values remain read-only until an accurate editor exists.
- Coalesce repeated edits to the same cell; restoring its original value removes that change. Inserting then deleting an unsaved row removes the draft operation.
- Offer Review Changes and Discard Changes. SQL preview is discoverable in review, with bound values represented safely; never inline passwords or imply preview text is the executed transport.
- Ctrl+S applies the active table tab only and consumes the shortcut so browser save does not fire. It is disabled during apply and when there are no edits or validation errors remain.
- Preserve drafts on failed apply. After success, clear only the applied draft and refresh/reconcile server values, including defaults and generated IDs.
- Closing, refreshing, paging, changing filters, or disconnecting must not silently lose drafts: preserve them by stable row identity or offer Apply / Discard / Cancel before the destructive navigation. Switching tabs can keep drafts in memory.

## Write safety and feedback

Only enable editing for a base table with a reliable primary/unique row key, sufficient permissions, and supported writable column types. Arbitrary SQL results, joins, views, generated columns, and tables without reliable row identity start read-only with a clear explanation. Do not update by visible row position or a broad predicate.

Go receives typed row operations and builds parameterized INSERT/UPDATE/DELETE statements with quoted identifiers. Revalidate schema and identifiers in Go. Retain original keys and original values for conflict detection; changing a key must still identify the original row. Use null-safe comparisons and type-aware predicates, backed by engine tests, so concurrent changes are detected. A matched/changed-row count must not be mistaken for a conflict merely because an update is a no-op.

For transactional tables in the initial edit implementation, apply the tab's batch atomically in a Go-owned transaction, checking all expected matches before commit. On validation/constraint/conflict failure, roll back the batch and retain the drafts with row-specific feedback. Report external transaction semantics honestly: triggers/nontransactional tables may prevent an absolute all-effects rollback guarantee. Initially block editing nontransactional tables rather than suggesting atomicity.

If commit acknowledgement is lost, show outcome unknown, keep the draft marked for reconciliation, and offer inspection/reload; never blindly retry. Explicit manual transaction mode is separate later work: applying a draft in that mode executes changes into the open transaction, and only a distinct Commit Transaction action commits it. Display that distinction visibly before enabling the mode.

## Delivery order and acceptance

M1 delivers the connection manager, workspace shell, explorer, read-only table grid, and SQL editor/results. Queued editing remains the next major data workflow once session/value correctness is validated; its M3 position records dependencies rather than lower product importance. Reserve UI space and state ownership now without simulating successful writes.

- [x] Record user-directed connection/explorer/editor navigation and queued-edit workflow.
- [x] Build the connection manager and connected workspace.
- [ ] Validate keyboard navigation, focus, resizing, and connection context.
- [ ] Implement table drafts, change indicators, review, and discard.
- [ ] Make Ctrl+S and Apply Changes use the same active-tab apply path.
- [ ] Prove no database changes occur while editing before apply.
- [ ] Test insert/update/delete coalescing, NULL, large values, and key changes.
- [ ] Test batch rollback, constraints, concurrent edits, stale schema, and unknown commit outcomes on both engines.
- [ ] Guard navigation and preserve drafts after errors without automatic replay.
- [ ] Keep manual SQL execution and explicit transaction commit distinct from draft application.

# First workflow: desktop test checklist

This build is ready for user testing of connection → explorer → table inspection/local text drafts → SQL execution. Drafts cannot be applied yet; Ctrl+S application, saved connections, history, pagination, and manual transaction controls remain unimplemented. Explicit SQL can modify data; use only the disposable fixtures below.

## Start

```powershell
Set-Location C:\projects\QueryLight
docker compose up -d --wait --wait-timeout 240
.\build\bin\querylight.exe
```

Alternatively use `wails dev` for the same application with reload. The connection manager's **MariaDB :13306** and **MySQL :13307** buttons fill the documented synthetic local credentials and disable TLS for those fixtures. Normal new connections default to verified TLS. Passwords and SQL text are not saved to disk.

## Connection and navigation

- [ ] Test Connection succeeds for each local preset and reports the expected version.
- [ ] A deliberately wrong password gives useful feedback without freezing the window.
- [ ] Connect opens the explorer and a SQL tab in querylight_test.
- [ ] Open several tables; search filters the explorer; refresh lists works.
- [ ] Open both servers and switch using the connection selector; each keeps its own tabs/results.
- [ ] Changing the explorer database does not silently change an existing SQL tab's session. The tab's database is visible above the editor; open a new SQL tab for the newly selected database.
- [ ] Click New SQL editor, then immediately change the explorer database while the editor opens. The new tab must retain the database requested at click time; subsequent new tabs use the new selection.

## Data inspection

- [x] Metadata-foundation build: grid_rows displayed its read-only explanation and fetched 500 rows on both engines in user screenshots (2026-10-15). The newer local-draft toolbar is covered by the steps below.
- [x] Open odd\`table name: toolbar explains a primary key is required for initial queued editing. User screenshots confirm the identifier-quoting fixture loads on both engines (2026-10-15).

- [ ] value_samples distinguishes NULL from empty text and preserves large integers/decimal trailing zeros.
- [ ] Click a cell to inspect its full transferred value; Copy value preserves it exactly. Binary and unsupported values are explicitly base64.
- [ ] grid_rows shows 500 rows and Result limit reached; scroll to the bottom.
- [ ] empty_table retains headers and says No rows returned.
- [ ] large_cells labels the 300000-byte payload unavailable and disables copying it as a complete value.
- [ ] temporal_edge_cases shows zero dates without normalization.
- [ ] odd\`table name opens correctly despite its embedded backtick.
- [ ] Arrow keys navigate the focused grid; horizontal scrolling reveals additional columns.

## Local table drafts

Repeat on both local presets. These steps do not apply drafts to the database.

- [ ] Open grid_rows, select the first label cell, click Edit local draft, enter multiline/Unicode text, then Queue draft. Expect a changed-cell dot/highlight, one pending row, and one changed cell.
- [ ] Edit the same cell again: counts stay at one. Restore `Synthetic row 1`: the draft disappears. Review Changes shows original and draft values without modifying the server.
- [ ] Open value_samples and edit nullable_text from NULL to empty text by unchecking NULL and leaving Draft value empty. Queue draft; confirm it displays `(empty)`. Edit it back to NULL; the draft disappears. A nonnullable text column must disable NULL.
- [ ] Primary keys, numbers, temporal/JSON/binary/generated and unavailable values cannot be drafted. Keyless tables and SQL results remain read-only. large_cells must not offer editing its unavailable payload.
- [ ] Queue a draft, switch SQL/table tabs, switch connections, and change explorer database. Return to the original table tab; its queued draft remains. Inspector input must be queued before switching.
- [ ] Try Refresh, Close table, and Disconnect with a queued draft. Keep working preserves it. Explicitly confirm discard/refresh and verify only that tab’s drafts are removed.
- [ ] Apply Changes remains disabled. Ctrl+S must not execute SQL, apply, or clear drafts. From a separate SQL tab, `SELECT label FROM querylight_test.grid_rows WHERE id=1;` still returns `Synthetic row 1` while its table draft is present.
- [ ] Close all SQL tabs while retaining a table draft, then close the native window. In the Windows Yes/No dialog, No preserves the draft and Yes exits without requiring disconnect. No must be the default selection. Open connections conservatively prompt even with no drafts.
- [ ] Discard Changes asks for confirmation, restores fetched values, and clears counts. Restart does not recover drafts; nothing is persisted.

Limits: 256 KiB UTF-8 per draft cell, 256 changed cells, 1 MiB serialized drafts per tab. Validation errors retain inspector input. This slice supports text/NULL updates only; no insert/delete/key editing, transactional apply, permission validation, conflict checking, or SQL preview exists yet.

Local-draft verification (2026-10-15): frontend typecheck/build, Go unit tests/vet, live regression on MariaDB 11.4.12-MariaDB-ubu2404 and MySQL 8.4.11, twelve Playwright tests, and Windows Wails packaging passed. Three draft-model tests cover coalescing/original retention, NULL/empty, unsupported values, key availability, Unicode and bounds. The new Edge workflow test covers local changes, tab retention, review/discard, refresh/close/disconnect guards, disabled Apply, and Ctrl+S making no StartExecution calls. Browser screenshot inspected at 1280×820. Live Go tests validate the backend, not frontend draft interactions; the manual steps above remain pending, including unchanged server values and native close behavior.

Native-close bug reported and fixed (2026-10-15): Windows showed Yes/No, but the guard compared the result against a custom caption, preventing exit even after Yes. Wails' installed Windows implementation confirms native canonical responses. The guard now accepts only Yes and defaults to No. Frontend build, Go tests (including native dialog response/default regressions), go vet, Windows packaging, and diff checks passed. Desktop retest of Yes/No with an open connection remains pending; database/browser suites were not repeated for this dialog-only follow-up.

## SQL editor

**Rows** mode returns a grid. **Statement** mode reports affected rows and discards result grids. Neither mode is a read-only safeguard. Choose the mode before running; SQL is executed once, never retried to discover its result type. Ctrl+Enter runs the entire editor contents as one statement; highlighted selection execution is not implemented.

Run these individually in Rows mode:

```sql
SELECT * FROM value_samples ORDER BY id;
SELECT 1 AS duplicate_name, 2 AS duplicate_name;
SELECT REPEAT('x', 100000) AS payload FROM grid_rows LIMIT 100;
```

- [ ] Run and Ctrl+Enter produce the same behavior.
- [ ] Duplicate names retain both columns, and the large result is capped with truncation feedback.
- [ ] A syntax error leaves SQL text intact and the editor session usable.
- [ ] Run `DO 1;` in Rows mode: the notice explains that SQL already ran and Statement mode is needed for affected-row counts. In Statement mode, the execution notice remains visible alongside the affected-row feedback.
- [ ] SELECT SLEEP(15) keeps the UI responsive; Cancel settles the operation and marks the lost editor session honestly. Its text remains available. New SQL session opens a fresh tab.
- [ ] Switching tabs during a query does not display its result in another tab.

To test writes without changing seeded tables, run each statement separately in **Statement** mode in the same SQL tab:

```sql
CREATE TEMPORARY TABLE querylight_manual_test (id INT PRIMARY KEY, label VARCHAR(30));
INSERT INTO querylight_manual_test VALUES (1, 'one'), (2, 'two');
UPDATE querylight_manual_test SET label = 'updated' WHERE id = 1;
DELETE FROM querylight_manual_test WHERE id = 2;
```

- [ ] The insert reports 2 affected rows; update/delete each report 1.
- [ ] Switching to Rows mode and running SELECT * FROM querylight_manual_test shows the remaining updated row.
- [ ] The temporary table is unavailable in a separate SQL session and disappears when its editor closes.

## Desktop behavior

- [ ] Layout works at 1280×820 and the minimum 900×600, including Windows display scaling.
- [ ] Resize the editor/results divider using the editor pane's lower-right resize handle.
- [ ] Window icon, minimize/restore, and clipboard work in WebView2.
- [ ] Closing a SQL tab asks before discarding text; Keep working preserves it.
- [ ] Disconnect asks before discarding tabs and cancels owned work.
- [ ] Closing the application with SQL tabs/running work prompts; No leaves it open, Yes exits. Windows native question dialogs use Yes/No rather than custom captions.
- [ ] Restart shows a clean connection manager without persisted credentials/text.

Record engine, screen scaling/window size, SQL used, expected/actual behavior, and any error message when reporting a failure. Avoid including real credentials.

User desktop evidence, 2026-10-15: Rows-mode `DO 1` displays the already-executed notice; Statement-mode `DO 1` completes with zero affected rows and its execution notice. Existing SQL 1 retains querylight_test and successfully selects value_samples after the explorer switches to information_schema; new SQL 6 uses information_schema. Screenshots establish database-context behavior but do not independently establish the precise timing of a race reproduction. Native write/cancel/clipboard/close acceptance remains pending.

Follow-up on this PC: elevated Docker inspection confirmed both fixture services healthy (the earlier API denial was sandbox access). Live integration was rerun successfully on MariaDB 11.4.12-MariaDB-ubu2404 and MySQL 8.4.11, including the new primary-key/column metadata and keyless read-only checks. Unit checks and go vet passed for the table metadata foundation. Views/non-InnoDB structural rejection has unit coverage; live view/generated-column/nontransactional editing fixtures and all actual draft application remain future work.

Table metadata delivery checks: frontend typecheck/build, all eight Edge mock tests, and Windows `wails build` also passed. Wails regenerated the DTO bindings; its generated models contain whitespace-only lines reported by `git diff --check`. These were left generator-owned per contributor instructions. User desktop screenshots confirm the new read-only explanations and successful previews on both engines: grid_rows fetched 500 rows (MariaDB 20 ms, MySQL 24 ms); the keyless identifier fixture fetched one row (MariaDB 12 ms, MySQL 16 ms). These displayed query durations are not scrolling or overall UI performance measurements. Actual queued editing remains unimplemented.

## Automated verification

```powershell
npm run build --prefix frontend
go test ./...
go vet ./...
docker compose up -d --wait --wait-timeout 240
go test -tags integration ./internal/application -v
npm run test:e2e --prefix frontend
wails build
```

The tagged Go integration tests use only localhost fixture ports (including QUERYLIGHT_MARIADB_PORT/QUERYLIGHT_MYSQL_PORT overrides), synthetic credentials, and temporary write tables. Browser tests use Microsoft Edge and a deliberate Wails mock; they verify UI behavior, not the native bridge. The desktop checklist above validates that final bridge/WebView2 path. Install Edge for these browser checks; no separate Playwright browser download is required with the current channel setting.

Verified 2026-10-15: Go unit checks/vet, live integration on MariaDB 11.4.12 and MySQL 8.4.11, and six Edge UI tests passed. Manual WebView2 testing and trusted-certificate TLS fixtures remain pending. A race-detector run has not been performed.

New-PC verification, 2026-10-15 (C:\projects\QueryLight): locked frontend dependency installation, frontend typecheck/build, Go unit tests, go vet, all eight Edge mock tests, and the Windows Wails package build passed. The two additional browser regressions cover database selection during editor creation and notices for zero-column results. Node reports v24.21.0 on this PC. Docker's Linux-engine API returned permission denied during inspection; live integration tests were not rerun here while Docker initializes. Earlier engine results above are historical evidence, not current-PC validation. Native cancellation, Statement writes, connection switching, clipboard, and window-close acceptance remain pending; user reports the UI, table loading, SELECT execution, and oversized-cell labeling work.

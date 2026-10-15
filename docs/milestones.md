# Milestones and acceptance

Track completed and remaining tasks in the [build guide checklist](build-guide.md#delivery-checklist). The acceptance requirements below remain the milestone gates; a successful scaffold build does not satisfy M1.

## M0: foundation

Inspect scaffold, initialize local Git, document product/architecture/commands/decisions/contributor instructions, and supply branding assets. No substantial database implementation. Remote setup awaits hosting details.

## M1: first complete workflow

Implementation is ready for [desktop acceptance testing](testing-m1.md). Automated driver/integration and browser UI checks pass; native WebView2 behavior, trusted TLS, and performance measurements are still acceptance work. Table inspection is a first-page preview, without pagination.

1. Compatibility spike: select exact MariaDB/MySQL fixture versions; validate driver, lossless codec, TLS, dedicated sessions, and cancellation.
2. Temporary connection form/session registry: success, bad authentication, unreachable server, timeout, disconnect, reconnect. No password persistence yet.
3. Schema/table browser and bounded inspection: permitted objects, unusual identifiers, loading/empty/error states, explicit row limits.
4. SQL editor/execution: validate CodeMirror 6; execute one statement explicitly; show typed results or affected rows, duration, errors, and cancel.
5. Virtualized results and lifecycle hardening: wide tables, bounded memory, repeated runs, disconnect during execution, shutdown cleanup.

Acceptance requires recorded versions of both database engines. Fixtures cover NULL vs empty, signed/unsigned BIGINT boundaries, DECIMAL trailing zeros, fractional/zero dates, JSON, binary zero bytes, Unicode, duplicate column names, empty results, and oversized cells. Copying complete supported values preserves their payload. Demonstrate row/byte caps and clear clipping/truncation feedback.

Cancel a slow query while typing: UI stays usable, operation settles, lost session is marked. Late results cannot overwrite newer runs. Failed SQL preserves editor text. Explicit writes show affected rows; cancellation never claims rollback. Verify session variables stay on their editor connection.

Demonstrate responsive scrolling of a 500-row wide result on the recorded Windows machine; capture measurements before setting stricter performance budgets. Run wails dev and the packaged executable with WebView2. Mocks/frontend builds alone do not prove database compatibility or desktop behavior.

## M2: daily-use persistence

Saved connection metadata, OS-backed passwords, versioned preferences, opt-in history with retention/deletion, and unsaved query recovery. Test locked/missing secret storage and confirm no plaintext passwords in files/logs/history.

## M3: controlled editing

Prioritize the [queued table editing workflow](ui-design.md): local drafts, visible pending changes, review/discard, and Ctrl+S/Apply Changes. Ship transactional batch application and conflict feedback before broader manual transaction controls.

Local text/NULL update drafts, review/discard, per-tab counts, and navigation guards are implemented for desktop acceptance. Apply Changes is disabled; inserts/deletes, additional types, privilege/schema revalidation, conflicts, and transactional batch execution remain pending.

Explicit transactions and table editing with reliable row identity/conflict feedback. Test DDL implicit commits, connection loss, and uncertain write outcomes. SSH, export, and administration need separate scope decisions.

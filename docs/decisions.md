# Decisions

Adopted for the first implementation: go-sql-driver/mysql v1.9.3, CodeMirror 6 with MySQL syntax, and TanStack Vue Virtual (rows and columns). Live driver tests pass on both fixture engines; trusted-certificate TLS remains unvalidated. Operations use bounded polling initially. The application service owns the initial registry and metadata logic; dedicated packages can be extracted when needed. Rows/Statement execution mode is explicit and never causes automatic SQL replay.

UI direction (2026-10-15): connection manager followed by an explorer and tabbed data/SQL workspace. Queue table edits locally, then apply the active tab with Ctrl+S or Apply Changes. Keep SQL Run and explicit transaction Commit separate. Details and acceptance checklist: [ui-design.md](ui-design.md).

Initial local database fixture targets: MariaDB 11.4.12 and MySQL 8.4.11, using official Docker images in compose.yaml. Actual client compatibility, TLS, and supported-version policy still require validation.

Approved branding: the Q/light logo in assets/branding/querylight.svg (2026-10-15). Keep the current design for favicon and desktop/shortcut icons.

Settled: Querylight display name; querylight folder/repository/executable. Wails v2 + Go + Vue 3/TypeScript/Vite, with Vue single-file components. Framework selection is closed. Database ownership stays in Go behind a small typed interface. Bound results, virtualize the grid, preserve values accurately, and keep secrets outside ordinary settings/source control. Design for cancellation and transactions now.

Proposed: database/sql + go-sql-driver/mysql, CodeMirror 6, tagged string payloads, and the limits in architecture.md. Validate driver/editor/grid choices with small spikes. No virtualization or credential-store library selected yet.

Open: supported server version policy beyond the selected local fixtures, remote owner/provider/visibility, license, naming availability, later OS support. Windows is the first validation target. Keep local module querylight until hosting is decided. Defer tray integration until a background workflow justifies it.

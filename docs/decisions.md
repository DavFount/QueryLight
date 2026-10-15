# Decisions

Approved branding: the Q/light logo in assets/branding/querylight.svg (2026-10-15). Keep the current design for favicon and desktop/shortcut icons.

Settled: Querylight display name; querylight folder/repository/executable. Wails v2 + Go + Vue 3/TypeScript/Vite, with Vue single-file components. Framework selection is closed. Database ownership stays in Go behind a small typed interface. Bound results, virtualize the grid, preserve values accurately, and keep secrets outside ordinary settings/source control. Design for cancellation and transactions now.

Proposed: database/sql + go-sql-driver/mysql, CodeMirror 6, tagged string payloads, and the limits in architecture.md. Validate driver/editor/grid choices with small spikes. No virtualization or credential-store library selected yet.

Open: exact supported server versions, remote owner/provider/visibility, license, naming availability, later OS support. Windows is the first validation target. Keep local module querylight until hosting is decided. Defer tray integration until a background workflow justifies it.

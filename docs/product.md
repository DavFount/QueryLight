# Product scope

See [UI design and queued editing](ui-design.md) for the connection manager, connected workspace, explorer/editor navigation, and the core draft-edit → Ctrl+S/Apply Changes workflow. Table editing is a planned core daily-use feature, dependent on the first read-only milestone's session/value correctness.

Querylight is a focused desktop SQL manager for MariaDB/MySQL, initially for RedM server owners and developers. Borrow familiar workflows from HeidiSQL and TablePlus, with fewer competing controls. Naming availability has not been checked.

The first milestone is connect → browse schemas/tables → inspect bounded rows → execute one SQL statement and view results. SQL execution can change data; table browsing is read-only initially. Never execute automatically after reconnecting.

Use a schema sidebar, central editor/table tabs, and results below. Keep connect, run, cancel, and connection/schema context visible. Put occasional actions in contextual menus. Preserve editor text on errors. Distinguish loading, empty, failed, cancelled, successful, and truncated results. NULL must look different from empty text. Support keyboard navigation, visible focus, readable contrast, and non-color status indicators.

Defer table editing, SSH tunnels, exports, advanced administration, and server monitoring. Saved connections, preferences, and opt-in history follow the first complete workflow.

Recommendation: no tray behavior initially. Minimize goes to the taskbar; closing releases sessions and exits. Add a close guard for unsaved SQL/running work/open transactions when those features exist. An opt-in tray needs an intentional background workflow first.

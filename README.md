# Querylight

A focused MariaDB/MySQL desktop SQL manager built with Wails v2, Go, Vue 3, TypeScript, and Vite. Status: first connection/explorer/SQL workflow and local text drafts ready for desktop testing. Drafts cannot be applied to the database yet; persistence is later work. Naming availability is unchecked.

## Development

Start with the [architecture and build guide with completion checklist](docs/build-guide.md) for setup, implementation order, validation, and remaining work.

Run from the repository root:

~~~powershell
npm ci --prefix frontend
wails dev
npm run build --prefix frontend
go test ./...
go vet ./...
wails build
~~~

Build frontend before Go checks because Go embeds frontend/dist. Desktop output goes to build/bin. Do not scaffold again. Browser-only Vite cannot exercise the Go bridge.

Verified 2026-10-15: Windows amd64, Wails 2.16.0, Go 1.27.2, Node 24.22.0. User reports npm 11.19.1, WebView2 and Wails Doctor passing. go.mod declares Go 1.25.0; Wails is pinned to 2.16.0. Commit dependency lockfiles.

## Structure and plans

Start with the [desktop test checklist](docs/testing-m1.md) to try the first working build and reproduce automated checks.

The [UI design](docs/ui-design.md) specifies the connection manager, explorer/editor workspace, and queued table editing with Ctrl+S.

Eligible table previews support local text/NULL drafts with changed-cell markers, pending row counts, Review, and Discard. Apply Changes is disabled and Ctrl+S does not write. Drafts remain in memory per tab; refresh, close, disconnect, and native exit guard their loss. See the [local draft test steps](docs/testing-m1.md#local-table-drafts). Structural eligibility does not imply write permission.

For live local MariaDB and MySQL fixtures, see [database testing](docs/database-testing.md). Start both with `docker compose up -d --wait --wait-timeout 240`.

main.go configures the desktop; app.go exposes the typed application facade. internal/application owns database connections, editor sessions, metadata, and operations; internal/database contains value conversion and identifier quoting. frontend/src holds the Vue workspace, editor, grid, and bridge adapter; frontend/wailsjs holds generated bindings. build holds icons and packaging inputs.

Read [product](docs/product.md), [architecture](docs/architecture.md), [milestones](docs/milestones.md), [decisions](docs/decisions.md), [repository setup](docs/repository.md), and [branding](docs/branding.md).

An origin remote points to https://github.com/DavFount/QueryLight.git; ownership/visibility and naming still need confirmation. No license is selected. Never commit credentials, real connection profiles, SQL history, or database data.


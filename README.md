# Querylight

A focused MariaDB/MySQL desktop SQL manager built with Wails v2, Go, Vue 3, TypeScript, and Vite. Status: scaffold and planning; database features are not implemented. Naming availability is unchecked.

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

For live local MariaDB and MySQL fixtures, see [database testing](docs/database-testing.md). Start both with `docker compose up -d --wait --wait-timeout 240`.

main.go configures the desktop; app.go is the greeting demo bridge. frontend/src holds Vue components; frontend/wailsjs holds generated bindings. build holds icons and packaging inputs.

Read [product](docs/product.md), [architecture](docs/architecture.md), [milestones](docs/milestones.md), [decisions](docs/decisions.md), [repository setup](docs/repository.md), and [branding](docs/branding.md).

An origin remote points to https://github.com/DavFount/QueryLight.git; ownership/visibility and naming still need confirmation. No license is selected. Never commit credentials, real connection profiles, SQL history, or database data.


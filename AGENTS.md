# Querylight contributor instructions

- Read README.md and docs/product.md, architecture.md, milestones.md, and decisions.md before features; keep documentation current.
- Use docs/build-guide.md as the implementation/build checklist. Check items only with evidence; distinguish compilation from manual runtime/database validation. Preserve the approved Q/light logo baseline.
- Use Wails v2, Go, Vue 3, TypeScript, Vite, and Vue script/template/style components. Display Querylight; repository/executable querylight. Framework selection is settled.
- Keep database execution, credentials, sessions, transactions, and cancellation in Go. Expose small typed DTOs.
- Preserve values losslessly; never JSON-number BIGINT/DECIMAL, conflate NULL/empty, or normalize temporal/binary values silently. Mark truncation.
- Bound rows, serialized bytes, cells, concurrency, and retained data. Virtualization does not replace backend bounds.
- Use dedicated editor sessions, contexts, explicit cleanup, and honest lost-session/uncertain-write feedback. Never auto-retry user SQL.
- Quote identifiers, bind values, and use driver config APIs. Never log/store passwords or DSNs. SQL history must be opt-in.
- Never commit real connection profiles, secrets, SQL history, dumps, or results. OS secret storage holds passwords; settings hold references.
- Never hand-edit frontend/wailsjs; regenerate with Wails. Commit lockfiles and packaging inputs; ignore binaries and frontend/dist.
- For feature changes, build frontend first (`npm run build --prefix frontend`), then run `go test ./...` and `go vet ./...`. Go embeds frontend/dist. Run `wails build` for desktop/packaging changes. Report limitations accurately.
- Test codecs/lifecycle meaningfully and validate both MariaDB/MySQL with recorded versions. Never run destructive tests on user databases.
- Stay within the milestone. Provide loading/empty/error/cancelled states. Do not create/push a remote, choose a license, publish, or introduce tray persistence without agreed scope/details.

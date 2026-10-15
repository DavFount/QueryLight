# Local database testing

The root compose.yaml provisions two independent disposable servers: MariaDB 11.4.12 and MySQL 8.4.11. Patch tags are explicit; these are initial test targets, not a claim of completed Querylight compatibility. Official image documentation: [MariaDB](https://hub.docker.com/_/mariadb), [MySQL](https://hub.docker.com/_/mysql).

Use Docker Desktop with Linux containers. From the repository root:

Verified on 2026-10-15: Compose configuration passed, both containers became healthy, and authenticated TCP client queries returned versions 11.4.12-MariaDB-ubu2404 and 8.4.11. Both returned 1000 grid rows, fixture version 1, expected BIGINT/DECIMAL/binary values, a 300000-byte cell, and preserved zero dates. Servers were left running. This validates the fixtures through native clients, not the unimplemented Querylight driver/bridge.

```powershell
docker compose config --quiet
docker compose up -d --wait --wait-timeout 240
docker compose ps
```

To run just MariaDB, append `mariadb` to the up command. Images download on first use. Named volumes preserve data between restarts. Nothing starts automatically on Windows boot through this configuration.

| Setting | MariaDB | MySQL |
| --- | --- | --- |
| Host | 127.0.0.1 | 127.0.0.1 |
| Port | 13306 | 13307 |
| Database | querylight_test | querylight_test |
| Username | querylight | querylight |
| Password | querylight-local-only | querylight-local-only |

These are deliberately public, synthetic local test credentials, not saved user credentials. Root password is querylight-root-local-only; normal tests use the database-scoped querylight user. Ports bind only to loopback. Do not deploy this configuration publicly or reuse these passwords elsewhere. Verified TLS is a separate test task; this setup does not provision trusted certificates or prove TLS behavior.

For port conflicts, set `$env:QUERYLIGHT_MARIADB_PORT = '23306'` or `$env:QUERYLIGHT_MYSQL_PORT = '23307'` before running Compose; keep the same overrides for subsequent commands.

## Inspect and smoke-test

Open a client and enter the local test password at its prompt:

```powershell
docker compose exec mariadb mariadb --protocol=TCP -h 127.0.0.1 -u querylight -p querylight_test
docker compose exec mysql mysql --protocol=TCP -h 127.0.0.1 -u querylight -p querylight_test
```

Run on each engine:

```sql
SELECT VERSION(), DATABASE();
SHOW TABLES;
SELECT * FROM fixture_ready; -- one row, fixture_version = 1
SELECT COUNT(*) FROM grid_rows; -- 1000
SELECT signed_bigint, unsigned_bigint, exact_decimal, HEX(binary_value) FROM value_samples;
SELECT * FROM temporal_edge_cases;
SELECT OCTET_LENGTH(payload) FROM large_cells; -- 300000
SELECT 1 AS duplicate_name, 2 AS duplicate_name;
SELECT * FROM empty_table;
SELECT SLEEP(15); -- later cancellation test; exit/interrupt as needed
```

Fixtures cover NULL/empty text, signed and unsigned limits, exact decimal scale, fractional and zero dates, JSON, binary/bit values, Unicode, awkward identifiers, an empty table, 1000 rows, and a large cell. They do not yet constitute a full integration runner, wide-grid benchmark, invalid-text-byte matrix, permissions matrix, or TLS test suite. MySQL JSON may normalize its representation differently from MariaDB; preserve what the server returns.

## Stop, logs, and reset

```powershell
docker compose logs --tail 100 mariadb mysql
docker compose stop
docker compose start
docker compose down
```

`down` removes these containers/network but retains their volumes. Initialization scripts run only on an empty data volume; editing SQL does not update existing databases. To deliberately erase **both local fixture databases** and seed again:

```powershell
docker compose down --volumes
docker compose up -d --wait --wait-timeout 240
```

Do not reset volumes containing work you want to retain. SQL fixtures are narrowly exempted from .gitignore; real dumps/history remain ignored. Record server VERSION() and image digests when recording integration results; patch tags can be rebuilt upstream.

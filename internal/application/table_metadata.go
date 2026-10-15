package application

import (
	"context"
	"database/sql"
	"strings"
)

// Metadata is bounded independently of the result grid. It is descriptive;
// no write permission or future apply authorization is inferred from it.
func tableMetadata(ctx context.Context, conn *sql.Conn, schema, table string) (*TableMetadata, error) {
	m := &TableMetadata{PrimaryKey: []string{}, Columns: []TableColumn{}}
	var kind string
	if err := conn.QueryRowContext(ctx, "SELECT TABLE_TYPE,COALESCE(ENGINE,'') FROM information_schema.TABLES WHERE TABLE_SCHEMA=? AND TABLE_NAME=?", schema, table).Scan(&kind, &m.Engine); err != nil {
		return nil, err
	}
	rows, err := conn.QueryContext(ctx, "SELECT COLUMN_NAME,DATA_TYPE,IS_NULLABLE,EXTRA,COLUMN_KEY FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=? ORDER BY ORDINAL_POSITION LIMIT 257", schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c TableColumn
		var nullable, extra, key string
		if err = rows.Scan(&c.Name, &c.DataType, &nullable, &extra, &key); err != nil {
			return nil, err
		}
		if len(m.Columns) >= MaxColumns {
			return nil, errResultLimit
		}
		c.Nullable = nullable == "YES"
		upper := strings.ToUpper(extra)
		c.Generated = strings.Contains(upper, "VIRTUAL GENERATED") || strings.Contains(upper, "STORED GENERATED") || strings.Contains(upper, "PERSISTENT GENERATED")
		m.Columns = append(m.Columns, c)
		if key == "PRI" {
			m.PrimaryKey = append(m.PrimaryKey, c.Name)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	classifyTable(m, kind)
	return m, nil
}

func classifyTable(m *TableMetadata, kind string) {
	switch {
	case kind != "BASE TABLE":
		m.ReadOnlyReason = "Views are read-only."
	case !strings.EqualFold(m.Engine, "InnoDB"):
		m.ReadOnlyReason = "Queued editing initially requires an InnoDB table."
	case len(m.PrimaryKey) == 0:
		m.ReadOnlyReason = "Queued editing initially requires a primary key."
	default:
		m.EditCandidate = true
		m.ReadOnlyReason = "Queued editing is not implemented yet; this preview is read-only."
	}
}

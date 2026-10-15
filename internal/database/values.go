package database

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxCellBytes = 256 * 1024

// Value preserves integers/decimals as strings and unknown/invalid text as bytes.
func Value(value any, databaseType string) (kind, payload string, unavailable bool, byteLength int) {
	if value == nil {
		return "null", "", false, 0
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	case int64:
		raw = []byte(strconv.FormatInt(v, 10))
	case uint64:
		raw = []byte(strconv.FormatUint(v, 10))
	case float64:
		raw = []byte(strconv.FormatFloat(v, 'g', -1, 64))
	case float32:
		raw = []byte(strconv.FormatFloat(float64(v), 'g', -1, 32))
	default:
		raw = []byte(fmt.Sprint(v))
		return "unsupported", base64.StdEncoding.EncodeToString(raw), false, len(raw)
	}
	switch strings.TrimPrefix(strings.ToUpper(databaseType), "UNSIGNED ") {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT":
		kind = "integer"
	case "DECIMAL", "NEWDECIMAL", "NUMERIC":
		kind = "decimal"
	case "FLOAT", "DOUBLE":
		kind = "float"
	case "DATE", "DATETIME", "TIMESTAMP", "TIME", "YEAR":
		kind = "temporal"
	case "JSON":
		kind = "json"
	case "BINARY", "VARBINARY", "BLOB", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB", "BIT", "GEOMETRY":
		kind = "binary"
	case "CHAR", "VARCHAR", "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT", "ENUM", "SET":
		kind = "text"
	default:
		kind = "unsupported"
	}
	if kind != "binary" && !utf8.Valid(raw) {
		kind = "unsupported"
	}
	if len(raw) > MaxCellBytes {
		return kind, "", true, len(raw)
	}
	if kind == "binary" || kind == "unsupported" {
		return kind, base64.StdEncoding.EncodeToString(raw), false, len(raw)
	}
	return kind, string(raw), false, len(raw)
}
func QuoteIdentifier(name string) string { return "`" + strings.ReplaceAll(name, "`", "``") + "`" }

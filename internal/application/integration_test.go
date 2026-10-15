//go:build integration

package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These tests target only compose.yaml's synthetic local fixtures, never a DSN.
func TestLiveDatabaseWorkflow(t *testing.T) {
	for _, engine := range []struct {
		name, env string
		port      int
	}{{"MariaDB", "QUERYLIGHT_MARIADB_PORT", 13306}, {"MySQL", "QUERYLIGHT_MYSQL_PORT", 13307}} {
		t.Run(engine.name, func(t *testing.T) {
			port := engine.port
			if value := os.Getenv(engine.env); value != "" {
				var err error
				port, err = strconv.Atoi(value)
				if err != nil {
					t.Fatal(err)
				}
			}
			cfg := ConnectionConfig{Name: engine.name, Host: "127.0.0.1", Port: port, User: "querylight", Password: "querylight-local-only", Database: "querylight_test"}
			s := NewService(context.Background())
			defer s.Shutdown()
			bad := cfg
			bad.Password = "incorrect-local-test-password"
			if r := s.Connect(bad, true); r.Error == nil || r.Error.Code != "authentication" {
				t.Fatalf("bad authentication: %+v", r.Error)
			}
			r := s.Connect(cfg, false)
			if r.Error != nil {
				t.Fatal(r.Error)
			}
			id := r.Connection.ID
			t.Log("server", r.Connection.Version)
			meta := s.ListSchemas(id)
			if meta.Error != nil {
				t.Fatal(meta.Error)
			}
			found := false
			for _, schema := range meta.Schemas {
				if schema == "querylight_test" {
					found = true
				}
			}
			if !found {
				t.Fatal("fixture schema missing")
			}
			tables := s.ListTables(id, "querylight_test")
			if tables.Error != nil || len(tables.Tables) < 7 {
				t.Fatalf("tables: %+v", tables)
			}
			editor := s.OpenEditor(id, "querylight_test")
			if editor.Error != nil {
				t.Fatal(editor.Error)
			}
			run := func(text, mode string) OperationView {
				t.Helper()
				start := s.StartExecution(ExecutionRequest{EditorID: editor.ID, SQL: text, Mode: mode})
				if start.Error != nil {
					t.Fatal(start.Error)
				}
				op := await(t, s, start.ID)
				s.ReleaseOperation(start.ID)
				return op
			}
			success := func(text string) *QueryResult {
				t.Helper()
				op := run(text, "rows")
				if op.Error != nil {
					t.Fatal(op.Error)
				}
				return op.Result
			}
			values := success("SELECT * FROM value_samples ORDER BY id")
			checks := map[int]string{3: "-9223372036854775808", 4: "18446744073709551615", 5: "12345678901234567890.1234500000", 7: "2026-10-15 12:34:56.123456", 8: "-25:01:02.123456", 10: "AAH//kFC"}
			for column, want := range checks {
				if got := values.Rows[0][column].Value; got != want {
					t.Fatalf("column %d: got %q want %q", column, got, want)
				}
			}
			if values.Rows[0][12].Value != "Querylight — café 雪 🎮" || values.Rows[0][11].Value != "pQ==" {
				t.Fatal("Unicode or BIT payload changed")
			}
			var document map[string]any
			if err := json.Unmarshal([]byte(values.Rows[0][9].Value), &document); err != nil || document["large_integer"] != "18446744073709551615" {
				t.Fatal("JSON payload changed")
			}
			if values.Rows[0][1].Kind != "null" || values.Rows[0][2].Kind != "text" || values.Rows[0][2].Value != "" || values.Rows[0][4].Kind != "integer" || values.Rows[0][5].Kind != "decimal" || values.Rows[0][10].Kind != "binary" {
				t.Fatalf("wrong value tags: %+v", values.Rows[0])
			}
			if got := success("SELECT * FROM temporal_edge_cases").Rows[0][1].Value; got != "0000-00-00" {
				t.Fatal(got)
			}
			empty := success("SELECT * FROM empty_table")
			if len(empty.Rows) != 0 || len(empty.Columns) != 2 {
				t.Fatal("empty result lost column metadata")
			}
			duplicate := success("SELECT 1 AS same, 2 AS same")
			if len(duplicate.Columns) != 2 || duplicate.Rows[0][1].Value != "2" {
				t.Fatal("duplicate names lost data")
			}
			limit := success("SELECT * FROM grid_rows ORDER BY id")
			if len(limit.Rows) != 500 || !limit.Truncated {
				t.Fatal("row limit not enforced")
			}
			byteLimit := success("SELECT REPEAT('x',100000) AS payload FROM grid_rows LIMIT 100")
			encoded, _ := json.Marshal(byteLimit)
			if len(encoded) > MaxResultBytes || !byteLimit.Truncated || len(byteLimit.Rows) >= 100 {
				t.Fatal("serialized byte limit not enforced")
			}
			oversize := success("SELECT * FROM large_cells")
			if !oversize.Rows[0][1].Unavailable || oversize.Rows[0][1].Value != "" || oversize.Rows[0][1].ByteLength != 300000 {
				t.Fatal("oversized cell not marked")
			}
			inspected := s.StartInspection(id, "querylight_test", "odd`table name")
			if inspected.Error != nil {
				t.Fatal(inspected.Error)
			}
			quoted := await(t, s, inspected.ID)
			if quoted.Error != nil || len(quoted.Result.Rows) != 1 {
				t.Fatalf("identifier quoting: %+v", quoted.Error)
			}
			s.ReleaseOperation(inspected.ID)
			table := s.StartInspection(id, "querylight_test", "grid_rows")
			preview := await(t, s, table.ID)
			if preview.Error != nil || len(preview.Result.Rows) != 500 || !preview.Result.Truncated {
				t.Fatal("table preview limit")
			}
			s.ReleaseOperation(table.ID)
			invalid := run("SELEC invalid", "rows")
			if invalid.Error == nil || invalid.Error.Code != "sql" || invalid.Error.SessionLost {
				t.Fatal("syntax error should preserve session")
			}
			multi := run("SELECT 1; SELECT 2", "rows")
			if multi.Error == nil {
				t.Fatal("multiple statements accepted")
			}
			success("SELECT 1")
			columns := make([]string, 257)
			for i := range columns {
				columns[i] = "1 AS c" + strconv.Itoa(i)
			}
			tooWide := run("SELECT "+strings.Join(columns, ","), "rows")
			if tooWide.Error == nil || tooWide.Error.Code != "limit" || tooWide.Error.SessionLost {
				t.Fatal("column limit should not invalidate session")
			}
			if op := run("SET @querylight_session = 42", "statement"); op.Error != nil {
				t.Fatal(op.Error)
			}
			if got := success("SELECT @querylight_session").Rows[0][0].Value; got != "42" {
				t.Fatal("dedicated session state lost")
			}
			other := s.OpenEditor(id, "querylight_test")
			if other.Error != nil {
				t.Fatal(other.Error)
			}
			otherStart := s.StartExecution(ExecutionRequest{EditorID: other.ID, SQL: "SELECT @querylight_session", Mode: "rows"})
			otherResult := await(t, s, otherStart.ID)
			if otherResult.Error != nil || otherResult.Result.Rows[0][0].Kind != "null" {
				t.Fatal("session variables leaked")
			}
			s.ReleaseOperation(otherStart.ID)
			s.CloseEditor(other.ID)
			// Temporary table exists only on this disposable editor connection.
			if op := run("CREATE TEMPORARY TABLE querylight_write_test (id INT PRIMARY KEY, label VARCHAR(30))", "statement"); op.Error != nil {
				t.Fatal(op.Error)
			}
			insert := run("INSERT INTO querylight_write_test VALUES (1,'one'),(2,'two')", "statement")
			if insert.Error != nil || insert.Result.AffectedRows != "2" {
				t.Fatal("affected row count")
			}
			success("SELECT * FROM querylight_write_test")
			oldID := editor.ID
			s.CloseEditor(oldID)
			editor = s.OpenEditor(id, "querylight_test")
			if editor.Error != nil {
				t.Fatal(editor.Error)
			}
			if got := success("SELECT @querylight_session").Rows[0][0].Kind; got != "null" {
				t.Fatal("closed session reused without resetting")
			}
			slow := s.StartExecution(ExecutionRequest{EditorID: editor.ID, SQL: "SELECT SLEEP(15)", Mode: "rows"})
			if slow.Error != nil {
				t.Fatal(slow.Error)
			}
			if busy := s.StartExecution(ExecutionRequest{EditorID: editor.ID, SQL: "SELECT 1", Mode: "rows"}); busy.Error == nil || busy.Error.Code != "busy" {
				t.Fatal("concurrent session execution allowed")
			}
			time.Sleep(150 * time.Millisecond)
			s.CancelOperation(slow.ID)
			cancelled := await(t, s, slow.ID)
			if cancelled.Error == nil || cancelled.Error.Code != "cancelled" || !cancelled.Error.SessionLost || !cancelled.Error.OutcomeUnknown {
				t.Fatalf("cancel outcome: %+v", cancelled.Error)
			}
			s.ReleaseOperation(slow.ID)
			if r := s.StartExecution(ExecutionRequest{EditorID: editor.ID, SQL: "SELECT 1", Mode: "rows"}); r.Error == nil || !r.Error.SessionLost {
				t.Fatal("cancelled session was reused")
			}
			s.Disconnect(id)
			if s.ListTables(id, "querylight_test").Error.Code != "disconnected" {
				t.Fatal("disconnected handle reused")
			}
			t.Log(fmt.Sprintf("verified values, bounds, quoting, writes, isolation, cancellation on %s", engine.name))
		})
	}
}
func await(t *testing.T, s *Service, id string) OperationView {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		op := s.GetOperation(id)
		if op.Status != "running" {
			return op
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.CancelOperation(id)
	t.Fatal("operation did not settle")
	return OperationView{}
}

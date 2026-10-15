package application

import (
	"context"
	"strings"
	"testing"
)

func TestRejectInvalidRequestsWithoutConnecting(t *testing.T) {
	s := NewService(context.Background())
	defer s.Shutdown()
	if s.Connect(ConnectionConfig{Host: "", Port: 0}, false).Error.Code != "validation" {
		t.Fatal("invalid config accepted")
	}
	if s.StartExecution(ExecutionRequest{SQL: strings.Repeat("x", MaxSQLBytes+1), Mode: "rows"}).Error.Code != "validation" {
		t.Fatal("unbounded SQL accepted")
	}
	if s.StartExecution(ExecutionRequest{SQL: "SELECT 1", Mode: "unknown"}).Error.Code != "validation" {
		t.Fatal("unknown execution mode accepted")
	}
	if s.StartExecution(ExecutionRequest{SQL: "SELECT 1", Mode: "rows", EditorID: "missing"}).Error.Code != "session_lost" {
		t.Fatal("missing session accepted")
	}
	if s.ListSchemas("missing").Error.Code != "disconnected" {
		t.Fatal("missing connection accepted")
	}
	if s.GetOperation("missing").Error.Code != "expired" {
		t.Fatal("missing operation accepted")
	}
}
func TestCancelledServiceCannotConnect(t *testing.T) {
	s := NewService(context.Background())
	s.Shutdown()
	reply := s.Connect(ConnectionConfig{Host: "127.0.0.1", Port: 13306, User: "querylight"}, false)
	if reply.Error == nil || reply.Error.Code != "cancelled" {
		t.Fatalf("expected cancelled service, got %+v", reply.Error)
	}
}

func TestNativeCloseGuardIncludesConnectionsWithoutEditors(t *testing.T) {
	s := NewService(context.Background())
	defer s.cancel()
	if s.HasOpenWork() {
		t.Fatal("empty service should not prompt")
	}
	s.connections["table-workspace"] = &connection{}
	if !s.HasOpenWork() {
		t.Fatal("table-only workspace may contain frontend drafts and must prompt")
	}
	delete(s.connections, "table-workspace")
	if s.HasOpenWork() {
		t.Fatal("closed workspace should not prompt")
	}
}

package database

import (
	"strings"
	"testing"
)

func TestLosslessValues(t *testing.T) {
	tests := []struct {
		name            string
		raw             any
		typ, kind, want string
	}{
		{"null", nil, "VARCHAR", "null", ""},
		{"empty", []byte{}, "VARCHAR", "text", ""},
		{"unsigned", []byte("18446744073709551615"), "UNSIGNED BIGINT", "integer", "18446744073709551615"},
		{"signed", int64(-9223372036854775808), "BIGINT", "integer", "-9223372036854775808"},
		{"decimal", []byte("12345678901234567890.1234500000"), "DECIMAL", "decimal", "12345678901234567890.1234500000"},
		{"zero date", []byte("0000-00-00"), "DATE", "temporal", "0000-00-00"},
		{"json", []byte("{\"a\": 1}"), "JSON", "json", "{\"a\": 1}"},
		{"binary", []byte{0, 1, 255}, "VARBINARY", "binary", "AAH/"},
		{"invalid UTF8", []byte{255}, "VARCHAR", "unsupported", "/w=="},
		{"unknown bytes", []byte{0, 1, 255}, "VECTOR", "unsupported", "AAH/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, payload, unavailable, _ := Value(tt.raw, tt.typ)
			if kind != tt.kind || payload != tt.want || unavailable {
				t.Fatalf("got %q %q unavailable=%v", kind, payload, unavailable)
			}
		})
	}
}
func TestOversizedCellIsExplicitlyUnavailable(t *testing.T) {
	kind, payload, unavailable, length := Value(strings.Repeat("x", MaxCellBytes+1), "LONGTEXT")
	if kind != "text" || payload != "" || !unavailable || length != MaxCellBytes+1 {
		t.Fatal("oversized cell must not masquerade as a complete empty value")
	}
}
func TestQuoteIdentifier(t *testing.T) {
	if got := QuoteIdentifier("odd`table name"); got != "`odd``table name`" {
		t.Fatal(got)
	}
}

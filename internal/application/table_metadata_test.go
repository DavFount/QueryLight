package application

import (
	"strings"
	"testing"
)

func TestTableStructuralEligibility(t *testing.T) {
	for _, tc := range []struct {
		name, kind, engine string
		keys               []string
		candidate          bool
		reason             string
	}{
		{"keyed InnoDB", "BASE TABLE", "InnoDB", []string{"id"}, true, "not implemented"},
		{"composite key", "BASE TABLE", "innodb", []string{"owner", "id"}, true, "not implemented"},
		{"keyless", "BASE TABLE", "InnoDB", nil, false, "primary key"},
		{"nontransactional", "BASE TABLE", "MyISAM", []string{"id"}, false, "InnoDB"},
		{"view", "VIEW", "", nil, false, "Views"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &TableMetadata{Engine: tc.engine, PrimaryKey: tc.keys}
			classifyTable(m, tc.kind)
			if m.EditCandidate != tc.candidate || !strings.Contains(m.ReadOnlyReason, tc.reason) {
				t.Fatalf("unexpected eligibility: %+v", m)
			}
		})
	}
}

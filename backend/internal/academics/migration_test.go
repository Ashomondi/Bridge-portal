package academics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAcademicMigrationCoversAllTenantTables is a schema guard: it verifies the
// owned migration creates all ten Release 1 academic tables and forces RLS on
// each. It runs without PostgreSQL, so it also protects the contract in CI
// environments that do not start the database container.
func TestAcademicMigrationCoversAllTenantTables(t *testing.T) {
	upPath := filepath.Join("..", "..", "..", "migrations", "000003_r1_academics.up.sql")
	downPath := filepath.Join("..", "..", "..", "migrations", "000003_r1_academics.down.sql")

	upBytes, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downBytes, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	up := string(upBytes)
	down := string(downBytes)

	tables := []string{
		"subjects",
		"subject_teacher_assignments",
		"grading_scale_versions",
		"timetable_slots",
		"calendar_events",
		"attendance_daily",
		"student_marks",
		"report_comments",
		"result_releases",
		"released_marks_snapshots",
	}
	for _, table := range tables {
		if !strings.Contains(up, "CREATE TABLE "+table+" (") {
			t.Fatalf("up migration is missing CREATE TABLE %s", table)
		}
		if !strings.Contains(down, "DROP TABLE IF EXISTS "+table+" CASCADE;") {
			t.Fatalf("down migration is missing DROP TABLE IF EXISTS %s", table)
		}
	}
	for _, required := range []string{
		"ENABLE ROW LEVEL SECURITY",
		"FORCE ROW LEVEL SECURITY",
		"current_school_id()",
		"tenant_isolation",
		"UNIQUE (school_id, student_id, attendance_date)",
		"UNIQUE (school_id, student_id, term_id, subject_id)",
		"CHECK (cbc_level IN ('EE', 'ME', 'AE', 'BE'))",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("up migration is missing required fragment %q", required)
		}
	}
}

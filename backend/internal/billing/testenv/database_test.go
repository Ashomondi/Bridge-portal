package testenv

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDatabaseRequiredInCI(t *testing.T) {
	if os.Getenv("DEV05_MISSING_DATABASE_CHILD") == "1" {
		New(t, false)
		return
	}
	t.Setenv("CI", "true")
	t.Setenv("DEV05_TEST_DATABASE_URL", "")
	t.Setenv("DEV05_MISSING_DATABASE_CHILD", "1")
	cmd := exec.Command(os.Args[0], "-test.run=^TestDatabaseRequiredInCI$", "-test.v")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "DEV05_TEST_DATABASE_URL must be set in CI") || strings.Contains(string(output), "--- SKIP") {
		t.Fatalf("missing database did not fail in CI: %v\n%s", err, output)
	}
}

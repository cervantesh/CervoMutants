package doctor

import "testing"

func TestWarningChecksDoNotFailDoctor(t *testing.T) {
	check := warning("windows-onedrive", "workspace is under OneDrive")
	if !check.OK {
		t.Fatal("warning checks should not fail doctor")
	}
	if check.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", check.Severity)
	}
}

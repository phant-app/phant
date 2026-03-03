package setup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAutoPrependFromIni(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "99-phant.ini")
	content := "; comment\nauto_prepend_file = \"/home/test/.config/phant/php/phant_prepend.php\"\n"
	if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}

	got := readAutoPrependFromIni(iniPath)
	want := "/home/test/.config/phant/php/phant_prepend.php"
	if got != want {
		t.Fatalf("readAutoPrependFromIni(...) = %q, want %q", got, want)
	}
}

func TestReadAutoPrependFromIni_EmptyWhenMissing(t *testing.T) {
	got := readAutoPrependFromIni(filepath.Join(t.TempDir(), "missing.ini"))
	if got != "" {
		t.Fatalf("readAutoPrependFromIni(...) = %q, want empty string", got)
	}
}

func TestApplyValetLinuxRemediation_RequiresConfirmation(t *testing.T) {
	result := ApplyValetLinuxRemediation(context.Background(), false)

	if runtime := result.Supported; !runtime {
		t.Skip("linux-only remediation flow")
	}

	if result.Applied {
		t.Fatalf("ApplyValetLinuxRemediation(...).Applied = true, want false without confirmation")
	}

	if result.Confirmed {
		t.Fatalf("ApplyValetLinuxRemediation(...).Confirmed = true, want false")
	}

	if result.Message == "" {
		t.Fatalf("ApplyValetLinuxRemediation(...) message should explain confirmation requirement")
	}
}

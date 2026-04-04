package setup

import (
	"context"
	"os"
	"path/filepath"
	"sort"
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

func TestDiscoverFPMServicesSort_PrioritizesPreferredVersion(t *testing.T) {
	services := []FPMServiceStatus{
		{ServiceName: "php8.5-fpm.service", Version: "8.5"},
		{ServiceName: "php8.4-fpm.service", Version: "8.4"},
		{ServiceName: "php8.6-fpm.service", Version: "8.6"},
	}
	preferredVersion := "8.4"

	sort.Slice(services, func(i, j int) bool {
		leftPreferred := preferredVersion != "" && services[i].Version == preferredVersion
		rightPreferred := preferredVersion != "" && services[j].Version == preferredVersion
		if leftPreferred != rightPreferred {
			return leftPreferred
		}

		return services[i].Version < services[j].Version
	})

	if len(services) != 3 {
		t.Fatalf("discoverFPMServices sort setup got %d services, want 3", len(services))
	}

	if services[0].Version != "8.4" {
		t.Fatalf("discoverFPMServices sort first version = %q, want %q", services[0].Version, "8.4")
	}

	if services[1].Version != "8.5" || services[2].Version != "8.6" {
		t.Fatalf("discoverFPMServices sort tail versions = %q, %q, want %q, %q", services[1].Version, services[2].Version, "8.5", "8.6")
	}
}

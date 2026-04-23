package services

import (
	"testing"

	settingsinfra "phant/internal/infra/settings"
)

func TestLicenseServiceSaveAndGetLicenseKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	service := NewLicenseService()
	save := service.SaveLicenseKey("  PHANT-1234  ")
	if !save.Success {
		t.Fatalf("SaveLicenseKey(...) success = false, error=%q", save.Error)
	}
	if save.LicenseKey != "PHANT-1234" {
		t.Fatalf("SaveLicenseKey(...) licenseKey = %q, want %q", save.LicenseKey, "PHANT-1234")
	}

	got := service.GetLicenseKey()
	if got.Error != "" {
		t.Fatalf("GetLicenseKey(...) error = %q", got.Error)
	}
	if got.LicenseKey != "PHANT-1234" {
		t.Fatalf("GetLicenseKey(...) licenseKey = %q, want %q", got.LicenseKey, "PHANT-1234")
	}

	provider := settingsinfra.NewFileProvider()
	settings, err := provider.Load(t.Context())
	if err != nil {
		t.Fatalf("settings provider load error = %v", err)
	}
	if settings.LicenseKey != "PHANT-1234" {
		t.Fatalf("settings license key = %q, want %q", settings.LicenseKey, "PHANT-1234")
	}
}

func TestLicenseServiceClearLicenseKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	service := NewLicenseService()
	_ = service.SaveLicenseKey("PHANT-1234")

	save := service.SaveLicenseKey("   ")
	if !save.Success {
		t.Fatalf("SaveLicenseKey(clear) success = false, error=%q", save.Error)
	}
	if save.LicenseKey != "" {
		t.Fatalf("SaveLicenseKey(clear) licenseKey = %q, want empty", save.LicenseKey)
	}

	got := service.GetLicenseKey()
	if got.Error != "" {
		t.Fatalf("GetLicenseKey(...) error = %q", got.Error)
	}
	if got.LicenseKey != "" {
		t.Fatalf("GetLicenseKey(...) licenseKey = %q, want empty", got.LicenseKey)
	}
}

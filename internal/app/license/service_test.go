package license

import (
	"context"
	"testing"

	domainsettings "phant/internal/domain/settings"
)

func TestService_SaveAndGetKey(t *testing.T) {
	var stored domainsettings.AppSettings
	service := NewService(Dependencies{
		LoadSettings: func(context.Context) (domainsettings.AppSettings, error) {
			return stored, nil
		},
		SaveSettings: func(_ context.Context, settings domainsettings.AppSettings) error {
			stored = settings
			return nil
		},
	})

	save := service.SaveKey(context.Background(), "  PHANT-1234  ")
	if !save.Success || save.Error != "" {
		t.Fatalf("SaveKey(...) success=%v error=%q", save.Success, save.Error)
	}
	if stored.LicenseKey != "PHANT-1234" {
		t.Fatalf("stored key = %q, want %q", stored.LicenseKey, "PHANT-1234")
	}

	got := service.GetKey(context.Background())
	if got.Error != "" {
		t.Fatalf("GetKey(...) error = %q", got.Error)
	}
	if got.LicenseKey != "PHANT-1234" {
		t.Fatalf("GetKey(...) licenseKey = %q, want %q", got.LicenseKey, "PHANT-1234")
	}
}

package services

import (
	"context"
	"errors"
	"testing"
	"time"

	appphpmanager "phant/internal/app/phpmanager"
	domainphpmanager "phant/internal/domain/phpmanager"
)

func TestPHPServiceSwitchVersionTimesOut(t *testing.T) {
	original := phpServiceTimeouts
	phpServiceTimeouts.switchV = 25 * time.Millisecond
	t.Cleanup(func() {
		phpServiceTimeouts = original
	})

	deps := appphpmanager.Dependencies{
		Platform: func() string { return "linux" },
		SwitchVersion: func(ctx context.Context, _ string) domainphpmanager.ActionResult {
			<-ctx.Done()
			return domainphpmanager.ActionResult{Supported: true, Error: ctx.Err().Error()}
		},
	}

	svc := &PHPService{service: appphpmanager.NewService(deps)}
	result := svc.SwitchPHPVersion("8.3")
	if result.Error == "" {
		t.Fatalf("SwitchPHPVersion(...) expected timeout error")
	}
	if result.Error != context.DeadlineExceeded.Error() {
		t.Fatalf("SwitchPHPVersion(...) error = %q, want %q", result.Error, context.DeadlineExceeded.Error())
	}
}

func TestPHPServiceInstallUsesLongerTimeoutBudget(t *testing.T) {
	original := phpServiceTimeouts
	phpServiceTimeouts.install = 60 * time.Millisecond
	t.Cleanup(func() {
		phpServiceTimeouts = original
	})

	deps := appphpmanager.Dependencies{
		Platform: func() string { return "linux" },
		InstallVersion: func(ctx context.Context, _ string) domainphpmanager.ActionResult {
			select {
			case <-time.After(20 * time.Millisecond):
				return domainphpmanager.ActionResult{Success: true, Supported: true, Message: "ok"}
			case <-ctx.Done():
				return domainphpmanager.ActionResult{Supported: true, Error: ctx.Err().Error()}
			}
		},
	}

	svc := &PHPService{service: appphpmanager.NewService(deps)}
	result := svc.InstallPHPVersion("8.3")
	if !result.Success {
		t.Fatalf("InstallPHPVersion(...) success = false, want true; error=%q", result.Error)
	}
}

func TestPHPServicePassesThroughImmediateError(t *testing.T) {
	deps := appphpmanager.Dependencies{
		Platform: func() string { return "linux" },
		SwitchVersion: func(_ context.Context, _ string) domainphpmanager.ActionResult {
			return domainphpmanager.ActionResult{Supported: true, Error: errors.New("boom").Error()}
		},
	}

	svc := &PHPService{service: appphpmanager.NewService(deps)}
	result := svc.SwitchPHPVersion("8.3")
	if result.Error != "boom" {
		t.Fatalf("SwitchPHPVersion(...) error = %q, want %q", result.Error, "boom")
	}
}

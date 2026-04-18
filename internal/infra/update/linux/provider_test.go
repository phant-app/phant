package linux

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	goos      string
	lastName  string
	lastArgs  []string
	returnErr error
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	f.lastName = name
	f.lastArgs = append([]string{}, args...)
	return "", f.returnErr
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	return file, nil
}

func (f *fakeRunner) GOOS() string {
	return f.goos
}

func TestProviderInstallDownloadedStartsInstaller(t *testing.T) {
	updateFile := filepath.Join(t.TempDir(), "phant-update.AppImage")
	if err := os.WriteFile(updateFile, []byte("new-appimage"), 0o755); err != nil {
		t.Fatalf("write update file: %v", err)
	}

	runner := &fakeRunner{goos: "linux"}
	provider := NewProvider(runner)

	result := provider.InstallDownloaded(context.Background(), updateFile)
	if result.Error != "" {
		t.Fatalf("InstallDownloaded(...) error = %q", result.Error)
	}
	if !result.Installed {
		t.Fatalf("InstallDownloaded(...) installed = false, want true")
	}
	if runner.lastName != "nohup" {
		t.Fatalf("InstallDownloaded(...) runner name = %q, want %q", runner.lastName, "nohup")
	}
	if len(runner.lastArgs) < 2 || runner.lastArgs[0] != "sh" {
		t.Fatalf("InstallDownloaded(...) runner args = %v, expected [sh <script>]", runner.lastArgs)
	}
	scriptBytes, err := os.ReadFile(runner.lastArgs[1])
	if err != nil {
		t.Fatalf("read generated script: %v", err)
	}
	script := string(scriptBytes)
	if !strings.Contains(script, "mv ") || !strings.Contains(script, "nohup ") {
		t.Fatalf("InstallDownloaded(...) script missing expected commands: %s", script)
	}
}

func TestProviderInstallDownloadedRejectsNonLinux(t *testing.T) {
	updateFile := filepath.Join(t.TempDir(), "phant-update.AppImage")
	if err := os.WriteFile(updateFile, []byte("new-appimage"), 0o755); err != nil {
		t.Fatalf("write update file: %v", err)
	}

	runner := &fakeRunner{goos: "darwin"}
	provider := NewProvider(runner)

	result := provider.InstallDownloaded(context.Background(), updateFile)
	if result.Error == "" {
		t.Fatalf("InstallDownloaded(...) expected non-linux error")
	}
}

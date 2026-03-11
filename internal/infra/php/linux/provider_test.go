package linux

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	domainphpmanager "phant/internal/domain/phpmanager"
)

type fakeRunner struct {
	goos     string
	commands []string
	outputs  map[string]string
	errors   map[string]error
	paths    map[string]string
}

func newFakeRunner(goos string) *fakeRunner {
	return &fakeRunner{
		goos:    goos,
		outputs: make(map[string]string),
		errors:  make(map[string]error),
		paths:   make(map[string]string),
	}
}

func (f *fakeRunner) GOOS() string {
	return f.goos
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	key := commandKey(name, args...)
	f.commands = append(f.commands, key)
	if err, ok := f.errors[key]; ok {
		return "", err
	}
	if output, ok := f.outputs[key]; ok {
		return output, nil
	}
	return "", nil
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if path, ok := f.paths[file]; ok {
		return path, nil
	}
	return "", exec.ErrNotFound
}

func commandKey(name string, args ...string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}

func TestDiscoverVersionsLinux(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.outputs["php -v"] = "PHP 8.2.12 (cli)"
	runner.outputs["dpkg-query -W -f=${Package}\\n"] = "php8.1-cli\nphp8.2-cli\n"
	runner.outputs["apt-cache search --names-only ^php[0-9]\\.[0-9]-cli$"] = "php8.3-cli - command-line interpreter for the PHP scripting language\n"

	provider := NewProvider(runner)
	active, versions, err := provider.DiscoverVersions(context.Background())
	if err != nil {
		t.Fatalf("DiscoverVersions returned error: %v", err)
	}

	if active != "8.2" {
		t.Fatalf("active version mismatch: got %q", active)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	if versions[0].Version != "8.3" || versions[0].Installed {
		t.Fatalf("unexpected first version: %+v", versions[0])
	}
	if versions[1].Version != "8.2" || !versions[1].Installed || !versions[1].Active {
		t.Fatalf("unexpected second version: %+v", versions[1])
	}
	if versions[2].Version != "8.1" || !versions[2].Installed {
		t.Fatalf("unexpected third version: %+v", versions[2])
	}
}

func TestDiscoverVersionsUnsupportedPlatform(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("darwin")
	provider := NewProvider(runner)

	active, versions, err := provider.DiscoverVersions(context.Background())
	if err != nil {
		t.Fatalf("DiscoverVersions returned error: %v", err)
	}
	if active != "" {
		t.Fatalf("expected empty active version, got %q", active)
	}
	if len(versions) != 0 {
		t.Fatalf("expected no versions, got %d", len(versions))
	}
}

func TestInstallVersionUnsupportedPlatform(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("darwin")
	provider := NewProvider(runner)

	result := provider.InstallVersion(context.Background(), "8.3")
	if result.Supported {
		t.Fatalf("expected unsupported action result")
	}
}

func TestInstallVersionAptPermissionError(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.paths["apt-get"] = "/usr/bin/apt-get"
	runner.outputs["dpkg-query --version"] = "Debian dpkg-query"
	runner.errors["dpkg-query -W -f=${Status} php8.3-cli"] = errors.New("package not installed")
	runner.errors["apt-get install -y php8.3 php8.3-cli php8.3-fpm php8.3-common"] = errors.New("permission denied")

	provider := NewProvider(runner)
	result := provider.InstallVersion(context.Background(), "8.3")
	if !result.Supported {
		t.Fatalf("expected supported action")
	}
	if !result.RequiresPrivilege {
		t.Fatalf("expected RequiresPrivilege=true")
	}
	if len(result.SuggestedCommands) == 0 {
		t.Fatalf("expected suggested sudo commands")
	}
}

func TestInstallVersionUsesPkexecFallback(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.paths["apt-get"] = "/usr/bin/apt-get"
	runner.paths["pkexec"] = "/usr/bin/pkexec"
	runner.outputs["dpkg-query --version"] = "Debian dpkg-query"
	runner.errors["dpkg-query -W -f=${Status} php8.3-cli"] = errors.New("package not installed")
	runner.errors["apt-get install -y php8.3 php8.3-cli php8.3-fpm php8.3-common"] = errors.New("permission denied")

	provider := NewProvider(runner)
	result := provider.InstallVersion(context.Background(), "8.3")
	if !result.Success {
		t.Fatalf("expected install to succeed via pkexec fallback, got error: %s", result.Error)
	}
	if !strings.Contains(result.Message, "via pkexec") {
		t.Fatalf("expected message to indicate pkexec usage, got %q", result.Message)
	}
}

func TestSetExtensionStateUsesPkexecFallback(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.paths["phpenmod"] = "/usr/sbin/phpenmod"
	runner.paths["pkexec"] = "/usr/bin/pkexec"
	runner.outputs["dpkg-query -W -f=${Package}\\n"] = "php8.3-cli\n"
	runner.errors["phpenmod -v 8.3 -s ALL xdebug"] = errors.New("permission denied")

	provider := NewProvider(runner)
	result := provider.SetExtensionState(context.Background(), domainphpmanager.ExtensionToggleRequest{Name: "xdebug", Enabled: true})
	if !result.Success {
		t.Fatalf("SetExtensionState(...) success = false, want true; error=%q", result.Error)
	}
	if !strings.Contains(result.Message, "via pkexec") {
		t.Fatalf("SetExtensionState(...) message = %q, want pkexec indicator", result.Message)
	}
}

func TestSwitchVersionValidation(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	provider := NewProvider(runner)

	result := provider.SwitchVersion(context.Background(), "8")
	if result.Error == "" {
		t.Fatalf("expected validation error")
	}
}

func TestSwitchVersionDoesNotInvokeValet(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.paths["php8.3"] = "/usr/bin/php8.3"
	runner.paths["valet"] = "/usr/bin/valet"

	provider := NewProvider(runner)
	result := provider.SwitchVersion(context.Background(), "8.3")
	if !result.Success {
		t.Fatalf("SwitchVersion(...) success = false, want true; error=%q", result.Error)
	}
	if !strings.Contains(result.Message, "If you use Valet") {
		t.Fatalf("SwitchVersion(...) message = %q, want valet hint", result.Message)
	}
	for _, command := range runner.commands {
		if strings.HasPrefix(command, "valet use") {
			t.Fatalf("SwitchVersion(...) unexpectedly invoked valet command: %q", command)
		}
	}
}

func TestSwitchVersionSetsAlternativesUsingResolvedBinaryPath(t *testing.T) {
	t.Parallel()

	runner := newFakeRunner("linux")
	runner.paths["php8.3"] = "/opt/php/8.3/bin/php"

	provider := NewProvider(runner)
	result := provider.SwitchVersion(context.Background(), "8.3")
	if !result.Success {
		t.Fatalf("SwitchVersion(...) success = false, want true; error=%q", result.Error)
	}
	if len(runner.commands) == 0 {
		t.Fatalf("SwitchVersion(...) expected commands to be executed")
	}
	found := false
	for _, command := range runner.commands {
		if command == "update-alternatives --set php /opt/php/8.3/bin/php" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("SwitchVersion(...) did not call update-alternatives with resolved binary path; commands=%v", runner.commands)
	}
}

func TestParsePHPVersionFromOutput(t *testing.T) {
	t.Parallel()

	got := parsePHPVersionFromOutput("PHP 8.3.7 (cli) (built: May 10 2026)")
	if got != "8.3" {
		t.Fatalf("expected 8.3, got %q", got)
	}
}

func TestUniqueSortedVersions(t *testing.T) {
	t.Parallel()

	got := uniqueSortedVersions([]string{"8.1", "8.3", "8.2", "8.3"})
	want := []string{"8.3", "8.2", "8.1"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected version at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

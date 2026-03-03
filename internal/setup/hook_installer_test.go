package setup

import (
	"strings"
	"testing"
)

func TestParseLoadedConfigurationFile(t *testing.T) {
	output := "Configuration File (php.ini) Path: /etc/php/8.4/cli\nLoaded Configuration File:         /etc/php/8.4/cli/php.ini\nScan for additional .ini files in: /etc/php/8.4/cli/conf.d"
	got := parseLoadedConfigurationFile(output)
	want := "/etc/php/8.4/cli/php.ini"
	if got != want {
		t.Fatalf("parseLoadedConfigurationFile(...) = %q, want %q", got, want)
	}
}

func TestParseAdditionalINIPath(t *testing.T) {
	output := "Configuration File (php.ini) Path: /etc/php/8.4/cli\nLoaded Configuration File:         /etc/php/8.4/cli/php.ini\nScan for additional .ini files in: /etc/php/8.4/cli/conf.d"
	got := parseAdditionalINIPath(output)
	want := "/etc/php/8.4/cli/conf.d"
	if got != want {
		t.Fatalf("parseAdditionalINIPath(...) = %q, want %q", got, want)
	}
}

func TestBuildConfDContent(t *testing.T) {
	got := buildConfDContent("/tmp/phant_prepend.php")
	if !strings.Contains(got, "auto_prepend_file = \"/tmp/phant_prepend.php\"") {
		t.Fatalf("buildConfDContent(...) missing auto_prepend_file entry")
	}
}

func TestEnsureAutoPrependBlock_AppendsWhenMissing(t *testing.T) {
	original := "memory_limit=512M\n"
	next, changed := ensureAutoPrependBlock(original, "/tmp/phant_prepend.php")
	if !changed {
		t.Fatalf("ensureAutoPrependBlock(...) changed = false, want true")
	}
	if next == original {
		t.Fatalf("ensureAutoPrependBlock(...) returned unchanged content")
	}
	if !strings.Contains(next, phantMarkerBegin) || !strings.Contains(next, phantMarkerEnd) {
		t.Fatalf("ensureAutoPrependBlock(...) missing markers")
	}
}

func TestEnsureAutoPrependBlock_ReplacesExistingBlock(t *testing.T) {
	original := "memory_limit=512M\n\n" +
		phantMarkerBegin + "\nauto_prepend_file = \"/old/path.php\"\n" +
		phantMarkerEnd + "\n"

	next, changed := ensureAutoPrependBlock(original, "/new/path.php")
	if !changed {
		t.Fatalf("ensureAutoPrependBlock(...) changed = false, want true")
	}
	if strings.Contains(next, "/old/path.php") {
		t.Fatalf("ensureAutoPrependBlock(...) still contains old path")
	}
	if !strings.Contains(next, "/new/path.php") {
		t.Fatalf("ensureAutoPrependBlock(...) missing new path")
	}
}

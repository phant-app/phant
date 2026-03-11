package setup

import "testing"

func TestParseValetLinksOutput(t *testing.T) {
	t.Parallel()

	input := `
+---------+-----+------------------+------------------------+-----+
| Site    | SSL | URL              | Path                   | PHP |
+---------+-----+------------------+------------------------+-----+
| blog    |     | http://blog.test | /home/ronald/code/blog | 8.4 |
| shop    | X   | https://shop.test| /home/ronald/code/shop | 8.3 |
+---------+-----+------------------+------------------------+-----+
`

	sites := parseValetLinksOutput(input)
	if got, want := len(sites), 2; got != want {
		t.Fatalf("parseValetLinksOutput(...) length = %v, want %v", got, want)
	}

	if got, want := sites[0].Name, "blog"; got != want {
		t.Fatalf("parseValetLinksOutput(...) first site name = %q, want %q", got, want)
	}
	if got, want := sites[0].Path, "/home/ronald/code/blog"; got != want {
		t.Fatalf("parseValetLinksOutput(...) first site path = %q, want %q", got, want)
	}
	if got, want := sites[0].URL, "http://blog.test"; got != want {
		t.Fatalf("parseValetLinksOutput(...) first site url = %q, want %q", got, want)
	}
	if got, want := sites[0].PHPVersion, "8.4"; got != want {
		t.Fatalf("parseValetLinksOutput(...) first site phpVersion = %q, want %q", got, want)
	}

	if got, want := sites[1].Name, "shop"; got != want {
		t.Fatalf("parseValetLinksOutput(...) second site name = %q, want %q", got, want)
	}
	if got, want := sites[1].Path, "/home/ronald/code/shop"; got != want {
		t.Fatalf("parseValetLinksOutput(...) second site path = %q, want %q", got, want)
	}
	if !sites[1].IsSecure {
		t.Fatalf("parseValetLinksOutput(...) second site isSecure = false, want true")
	}
}

func TestParseValetLinksOutput_Empty(t *testing.T) {
	t.Parallel()

	sites := parseValetLinksOutput("   \n")
	if sites != nil {
		t.Fatalf("parseValetLinksOutput(...) = %#v, want nil for empty output", sites)
	}
}

func TestParseValetLinksOutput_IgnoresMalformedRows(t *testing.T) {
	t.Parallel()

	input := `
| Site | Path |
| foo |
| bar | /home/ronald/code/bar |
not-a-table-line
`

	sites := parseValetLinksOutput(input)
	if got, want := len(sites), 1; got != want {
		t.Fatalf("parseValetLinksOutput(...) length = %v, want %v", got, want)
	}
}

func TestParseValetPathsOutput_JSON(t *testing.T) {
	t.Parallel()

	input := "[\n  \"/home/ronald/.valet/Sites\",\n  \"/home/ronald/code\"\n]"
	paths := parseValetPathsOutput(input)

	if got, want := len(paths), 2; got != want {
		t.Fatalf("parseValetPathsOutput(...) length = %v, want %v", got, want)
	}

	if got, want := paths[0], "/home/ronald/.valet/Sites"; got != want {
		t.Fatalf("parseValetPathsOutput(...) first path = %q, want %q", got, want)
	}
}

func TestBuildValetSiteURL(t *testing.T) {
	t.Parallel()

	if got, want := buildValetSiteURL("blog", "test", false), "http://blog.test"; got != want {
		t.Fatalf("buildValetSiteURL(blog, test, false) = %q, want %q", got, want)
	}

	if got, want := buildValetSiteURL("blog", "test", true), "https://blog.test"; got != want {
		t.Fatalf("buildValetSiteURL(blog, test, true) = %q, want %q", got, want)
	}
}

func TestUnsupportedValetSitesResult(t *testing.T) {
	t.Parallel()

	result := unsupportedValetSitesResult("windows")
	if result.Supported {
		t.Fatalf("unsupportedValetSitesResult(windows).Supported = true, want false")
	}
	if got, want := result.OS, "windows"; got != want {
		t.Fatalf("unsupportedValetSitesResult(windows).OS = %q, want %q", got, want)
	}
	if got, want := result.Source, "links+paths"; got != want {
		t.Fatalf("unsupportedValetSitesResult(windows).Source = %q, want %q", got, want)
	}
	if got, want := len(result.ParkedDirectories), 0; got != want {
		t.Fatalf("unsupportedValetSitesResult(windows).ParkedDirectories length = %v, want %v", got, want)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("unsupportedValetSitesResult(windows).Warnings is empty, want warning text")
	}
}

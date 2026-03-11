package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type ValetSite struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	URL        string `json:"url"`
	IsSecure   bool   `json:"isSecure"`
	PHPVersion string `json:"phpVersion,omitempty"`
}

type ValetSitesResult struct {
	GeneratedAt       string      `json:"generatedAt"`
	Supported         bool        `json:"supported"`
	OS                string      `json:"os"`
	Source            string      `json:"source"`
	Sites             []ValetSite `json:"sites"`
	ParkedDirectories []string    `json:"parkedDirectories"`
	Warnings          []string    `json:"warnings"`
	Error             string      `json:"error"`
}

func DiscoverValetSites(ctx context.Context) ValetSitesResult {
	if runtime.GOOS != "linux" {
		return unsupportedValetSitesResult(runtime.GOOS)
	}

	result := ValetSitesResult{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Supported:   true,
		OS:          runtime.GOOS,
		Source:      "links+paths",
	}

	if !isCommandAvailable("valet") {
		result.Error = "valet CLI was not found in PATH"
		return result
	}

	linksOutput, linksErr := runCommand(ctx, "valet", "links")
	pathsOutput, pathsErr := runCommand(ctx, "valet", "paths")

	if linksErr != nil && pathsErr != nil {
		result.Error = fmt.Sprintf("valet links failed: %v; valet paths failed: %v", linksErr, pathsErr)
		return result
	}

	if linksErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("valet links failed: %v", linksErr))
	}

	if pathsErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("valet paths failed: %v", pathsErr))
	}

	links := parseValetLinksOutput(linksOutput)
	if len(links) == 0 {
		result.Warnings = append(result.Warnings, "No linked sites were returned by valet links.")
	}

	paths := parseValetPathsOutput(pathsOutput)
	if len(paths) == 0 {
		result.Warnings = append(result.Warnings, "No parked directories were returned by valet paths.")
	}

	tld := detectValetDomainTLD(ctx)
	globalPHPVersion := detectDefaultPHPVersion(ctx)
	for i := range links {
		if links[i].URL == "" {
			links[i].URL = buildValetSiteURL(links[i].Name, tld, links[i].IsSecure)
		}
		if links[i].PHPVersion == "" {
			links[i].PHPVersion = globalPHPVersion
		}
	}

	result.Sites = links
	result.ParkedDirectories = paths
	return result
}

func unsupportedValetSitesResult(goos string) ValetSitesResult {
	return ValetSitesResult{
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
		Supported:         false,
		OS:                goos,
		Source:            "links+paths",
		Sites:             []ValetSite{},
		ParkedDirectories: []string{},
		Warnings: []string{
			fmt.Sprintf("Valet sites discovery is not implemented for %s yet.", goos),
		},
	}
}

func detectDefaultPHPVersion(ctx context.Context) string {
	output, err := runCommand(ctx, "php", "-r", "echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION;")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(output)
}

func detectValetDomainTLD(ctx context.Context) string {
	output, err := runCommand(ctx, "valet", "domain")
	if err != nil {
		return "test"
	}

	tld := strings.TrimSpace(output)
	tld = strings.TrimPrefix(tld, ".")
	if tld == "" {
		return "test"
	}

	return tld
}

func parseValetLinksOutput(output string) []ValetSite {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	sites := make(map[string]ValetSite)
	headerIndexes := map[string]int{}
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || !strings.Contains(line, "|") || strings.HasPrefix(line, "+") {
			continue
		}

		columns := parseTableColumns(line)
		if len(columns) == 0 {
			continue
		}

		if looksLikeLinksHeader(columns) {
			headerIndexes = buildHeaderIndexes(columns)
			continue
		}

		if len(columns) < 2 {
			continue
		}

		name := getColumnByKey(columns, headerIndexes, "site", 0)
		path := getColumnByKey(columns, headerIndexes, "path", 1)
		url := getColumnByKey(columns, headerIndexes, "url", 2)
		phpVersion := getColumnByKey(columns, headerIndexes, "php", 3)
		sslValue := getColumnByKey(columns, headerIndexes, "ssl", -1)

		if name == "" || path == "" {
			continue
		}
		if strings.EqualFold(name, "site") || strings.EqualFold(path, "path") {
			continue
		}

		site := ValetSite{
			Name:       name,
			Path:       path,
			URL:        url,
			PHPVersion: phpVersion,
			IsSecure:   strings.TrimSpace(sslValue) != "" || strings.HasPrefix(strings.ToLower(url), "https://"),
		}
		if !site.IsSecure {
			site.IsSecure = isValetSiteSecure(name)
		}
		sites[name] = site
	}

	if len(sites) == 0 {
		return nil
	}

	resolved := make([]ValetSite, 0, len(sites))
	for _, site := range sites {
		resolved = append(resolved, site)
	}

	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].Name < resolved[j].Name
	})

	return resolved
}

func parsePipeColumns(line string) []string {
	raw := strings.Split(line, "|")
	columns := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		columns = append(columns, trimmed)
	}
	return columns
}

func parseTableColumns(line string) []string {
	parts := strings.Split(line, "|")
	if len(parts) < 3 {
		return nil
	}

	parts = parts[1 : len(parts)-1]
	columns := make([]string, 0, len(parts))
	for _, part := range parts {
		columns = append(columns, strings.TrimSpace(part))
	}

	return columns
}

func looksLikeLinksHeader(columns []string) bool {
	header := strings.ToLower(strings.Join(columns, "|"))
	return strings.Contains(header, "site") && strings.Contains(header, "path")
}

func buildHeaderIndexes(columns []string) map[string]int {
	indexes := make(map[string]int, len(columns))
	for i, column := range columns {
		key := strings.ToLower(strings.TrimSpace(column))
		if key == "" {
			continue
		}
		indexes[key] = i
	}
	return indexes
}

func getColumnByKey(columns []string, indexes map[string]int, key string, fallback int) string {
	if index, ok := indexes[key]; ok && index >= 0 && index < len(columns) {
		return strings.TrimSpace(columns[index])
	}
	if fallback >= 0 && fallback < len(columns) {
		return strings.TrimSpace(columns[fallback])
	}
	return ""
}

func parseValetPathsOutput(output string) []string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}

	var jsonPaths []string
	if err := json.Unmarshal([]byte(trimmed), &jsonPaths); err == nil {
		return compactStrings(jsonPaths)
	}

	lines := strings.Split(trimmed, "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		candidate := strings.TrimSpace(line)
		candidate = strings.Trim(candidate, "\",")
		if candidate == "" || candidate == "[" || candidate == "]" {
			continue
		}
		paths = append(paths, candidate)
	}

	return compactStrings(paths)
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	resolved := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		resolved = append(resolved, trimmed)
	}

	if len(resolved) == 0 {
		return nil
	}

	return resolved
}

func isValetSiteSecure(siteName string) bool {
	if siteName == "" {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	candidates := []string{
		filepath.Join(homeDir, ".valet", "Certificates", siteName+".crt"),
		filepath.Join(homeDir, ".config", "valet", "Certificates", siteName+".crt"),
	}

	for _, path := range candidates {
		if fileExists(path) {
			return true
		}
	}

	return false
}

func buildValetSiteURL(siteName string, tld string, secure bool) string {
	if siteName == "" {
		return ""
	}

	scheme := "http"
	if secure {
		scheme = "https"
	}

	if tld == "" {
		tld = "test"
	}

	return fmt.Sprintf("%s://%s.%s", scheme, siteName, tld)
}

package linux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	domainphpmanager "phant/internal/domain/phpmanager"
	"phant/internal/infra/system"
)

var versionPattern = regexp.MustCompile(`^\d+\.\d+$`)

type Provider struct {
	runner system.Runner
}

func NewProvider(runner system.Runner) *Provider {
	return &Provider{runner: runner}
}

func (p *Provider) Platform() string {
	return p.runner.GOOS()
}

func (p *Provider) DiscoverVersions(ctx context.Context) (string, []domainphpmanager.Version, error) {
	if p.Platform() != "linux" {
		return "", nil, nil
	}

	versionOutput, err := p.runner.Run(ctx, "php", "-v")
	if err != nil {
		return "", nil, fmt.Errorf("php -v failed: %w", err)
	}

	activeVersion := parsePHPVersionFromOutput(versionOutput)
	installed, installErr := p.discoverInstalledVersions(ctx)
	if installErr != nil {
		if activeVersion == "" {
			return "", nil, installErr
		}
		return activeVersion, []domainphpmanager.Version{{
			Version:   activeVersion,
			Installed: true,
			Active:    true,
		}}, nil
	}

	available, availableErr := p.discoverAvailableVersions(ctx)
	if availableErr != nil {
		available = nil
	}

	allVersions := make(map[string]struct{}, len(installed)+len(available)+1)
	for _, version := range installed {
		allVersions[version] = struct{}{}
	}
	for _, version := range available {
		allVersions[version] = struct{}{}
	}
	if activeVersion != "" {
		allVersions[activeVersion] = struct{}{}
	}

	if len(allVersions) == 0 {
		return activeVersion, nil, nil
	}

	installedSet := make(map[string]struct{}, len(installed))
	for _, version := range installed {
		installedSet[version] = struct{}{}
	}

	versions := make([]string, 0, len(allVersions))
	for version := range allVersions {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	result := make([]domainphpmanager.Version, 0, len(versions))
	for _, version := range versions {
		_, isInstalled := installedSet[version]
		result = append(result, domainphpmanager.Version{
			Version:   version,
			Installed: isInstalled,
			Active:    version == activeVersion,
		})
	}

	return activeVersion, result, nil
}

func (p *Provider) InstallVersion(ctx context.Context, version string) domainphpmanager.ActionResult {
	if p.Platform() != "linux" {
		return unsupportedActionResult("install", version)
	}

	requested := strings.TrimSpace(version)
	if !isValidVersion(requested) {
		return domainphpmanager.ActionResult{
			Supported: true,
			Version:   requested,
			Error:     "version must use major.minor format, for example 8.3",
		}
	}

	if !p.isAptBasedLinux(ctx) {
		return domainphpmanager.ActionResult{
			Supported: false,
			Version:   requested,
			Message:   "automatic install currently supports apt-based Linux distributions only",
		}
	}

	installed, checkErr := p.isVersionInstalled(ctx, requested)
	if checkErr == nil && installed {
		return domainphpmanager.ActionResult{
			Success:   true,
			Supported: true,
			Version:   requested,
			Message:   fmt.Sprintf("PHP %s is already installed", requested),
		}
	}

	packages := []string{
		fmt.Sprintf("php%s", requested),
		fmt.Sprintf("php%s-cli", requested),
		fmt.Sprintf("php%s-fpm", requested),
		fmt.Sprintf("php%s-common", requested),
	}
	args := append([]string{"install", "-y"}, packages...)
	commandText := "apt-get " + strings.Join(args, " ")

	usedPkexec, err := p.runWithPrivilegeFallback(ctx, "apt-get", args...)
	if err != nil {
		result := domainphpmanager.ActionResult{
			Supported: true,
			Version:   requested,
			Command:   commandText,
			Error:     fmt.Sprintf("PHP install failed: %v", err),
		}
		if requiresRootPrivileges(err) {
			result.RequiresPrivilege = true
			result.SuggestedCommands = []string{
				"sudo apt-get update",
				"sudo " + commandText,
			}
		}
		return result
	}
	message := fmt.Sprintf("PHP %s installed successfully", requested)
	if usedPkexec {
		message += " (via pkexec)"
	}

	return domainphpmanager.ActionResult{
		Success:   true,
		Supported: true,
		Version:   requested,
		Command:   commandText,
		Message:   message,
	}
}

func (p *Provider) SwitchVersion(ctx context.Context, version string) domainphpmanager.ActionResult {
	if p.Platform() != "linux" {
		return unsupportedActionResult("switch", version)
	}

	requested := strings.TrimSpace(version)
	if !isValidVersion(requested) {
		return domainphpmanager.ActionResult{
			Supported: true,
			Version:   requested,
			Error:     "version must use major.minor format, for example 8.3",
		}
	}

	binaryName := "php" + requested
	binaryPath, err := p.runner.LookPath(binaryName)
	if err != nil {
		return domainphpmanager.ActionResult{
			Supported: true,
			Version:   requested,
			Error:     fmt.Sprintf("PHP %s binary was not found in PATH", requested),
		}
	}

	args := []string{"--set", "php", binaryPath}
	commandText := "update-alternatives " + strings.Join(args, " ")
	usedPkexec, err := p.runWithPrivilegeFallback(ctx, "update-alternatives", args...)
	if err != nil {
		result := domainphpmanager.ActionResult{
			Supported: true,
			Version:   requested,
			Command:   commandText,
			Error:     fmt.Sprintf("failed to switch CLI PHP: %v", err),
		}
		if requiresRootPrivileges(err) {
			result.RequiresPrivilege = true
			result.SuggestedCommands = []string{"sudo " + commandText}
		}
		return result
	}

	message := fmt.Sprintf("PHP %s is now active", requested)
	if usedPkexec {
		message += " (via pkexec)"
	}
	if _, err := p.runner.LookPath("valet"); err == nil {
		message += fmt.Sprintf(". If you use Valet, run: valet use php@%s", requested)
	}

	return domainphpmanager.ActionResult{
		Success:   true,
		Supported: true,
		Version:   requested,
		Command:   commandText,
		Message:   message,
	}
}

func (p *Provider) DiscoverSettings(ctx context.Context) (domainphpmanager.IniSettings, error) {
	if p.Platform() != "linux" {
		return domainphpmanager.IniSettings{}, nil
	}

	output, err := p.runner.Run(ctx, "php", "-r", phpSettingsReadScript())
	if err != nil {
		return domainphpmanager.IniSettings{}, fmt.Errorf("failed to read php.ini settings: %w", err)
	}

	return parsePHPIniSettingsOutput(output), nil
}

func (p *Provider) DiscoverExtensions(ctx context.Context) ([]domainphpmanager.Extension, error) {
	if p.Platform() != "linux" {
		return nil, nil
	}

	enabledOutput, err := p.runner.Run(ctx, "php", "-m")
	if err != nil {
		return nil, fmt.Errorf("php -m failed: %w", err)
	}

	enabledSet := parsePHPExtensionsOutput(enabledOutput)
	availableMap, discoverErr := discoverAvailableExtensionINIFiles()
	if discoverErr != nil {
		return nil, discoverErr
	}

	namesSet := make(map[string]struct{}, len(enabledSet)+len(availableMap))
	for name := range enabledSet {
		namesSet[name] = struct{}{}
	}
	for name := range availableMap {
		namesSet[name] = struct{}{}
	}

	names := make([]string, 0, len(namesSet))
	for name := range namesSet {
		names = append(names, name)
	}
	sort.Strings(names)

	extensions := make([]domainphpmanager.Extension, 0, len(names))
	for _, name := range names {
		iniPath, exists := availableMap[name]
		_, enabled := enabledSet[name]
		extensions = append(extensions, domainphpmanager.Extension{
			Name:      name,
			Enabled:   enabled,
			Scope:     "linux",
			INIPath:   iniPath,
			INIExists: exists,
		})
	}

	return extensions, nil
}

func (p *Provider) UpdateSettings(ctx context.Context, request domainphpmanager.IniSettingsUpdateRequest) domainphpmanager.ActionResult {
	if p.Platform() != "linux" {
		return unsupportedActionResult("update-settings", "")
	}

	content, contentErr := buildManagedPHPSettingsINI(request)
	if contentErr != nil {
		return domainphpmanager.ActionResult{Supported: true, Error: contentErr.Error()}
	}

	cliINIOutput, err := p.runner.Run(ctx, "php", "--ini")
	if err != nil {
		return domainphpmanager.ActionResult{Supported: true, Error: fmt.Sprintf("php --ini failed: %v", err)}
	}

	cliConfDPath := parseAdditionalINIPath(cliINIOutput)
	if cliConfDPath == "" || strings.EqualFold(cliConfDPath, "(none)") {
		return domainphpmanager.ActionResult{Supported: true, Error: "unable to detect CLI conf.d directory"}
	}

	targets := []string{filepath.Join(cliConfDPath, "99-phant-settings.ini")}
	services := discoverFPMServiceTargets()
	for _, service := range services {
		targets = append(targets, filepath.Join(service.ConfDPath, "99-phant-settings.ini"))
	}

	hasWriteFailure := false
	requiresPrivilege := false
	suggested := make([]string, 0, len(targets))
	usedPkexecForWrites := false
	for _, target := range targets {
		usedPkexec, writeErr := p.writeManagedINI(ctx, target, content)
		if usedPkexec {
			usedPkexecForWrites = true
		}
		if writeErr != nil {
			hasWriteFailure = true
			if isPermissionError(writeErr) || strings.Contains(strings.ToLower(writeErr.Error()), "pkexec") {
				requiresPrivilege = true
				suggested = append(suggested, buildLinuxWriteManagedINICommand(target, content))
			}
		}
	}

	if hasWriteFailure {
		return domainphpmanager.ActionResult{
			Supported:         true,
			RequiresPrivilege: requiresPrivilege,
			SuggestedCommands: uniqueStrings(suggested),
			Message:           "one or more php.ini targets failed to update",
			Error:             "failed to apply PHP settings to all targets",
		}
	}

	usedPkexecForRestart := false
	for _, service := range services {
		usedPkexec, restartErr := p.restartFPMService(ctx, service.ServiceName)
		if usedPkexec {
			usedPkexecForRestart = true
		}
		if restartErr != nil {
			return domainphpmanager.ActionResult{
				Supported:         true,
				RequiresPrivilege: true,
				SuggestedCommands: []string{service.RestartCommand},
				Error:             fmt.Sprintf("settings updated, but failed to restart %s: %v", service.ServiceName, restartErr),
				Message:           "settings updated, but one or more PHP-FPM services require manual restart",
			}
		}
	}

	message := "PHP settings updated for CLI and discovered FPM services"
	if usedPkexecForWrites || usedPkexecForRestart {
		message += " (via pkexec)"
	}

	return domainphpmanager.ActionResult{Success: true, Supported: true, Message: message}
}

func (p *Provider) SetExtensionState(ctx context.Context, request domainphpmanager.ExtensionToggleRequest) domainphpmanager.ActionResult {
	if p.Platform() != "linux" {
		return unsupportedActionResult("toggle-extension", request.Name)
	}

	extensionName := normalizeExtensionName(request.Name)
	if extensionName == "" {
		return domainphpmanager.ActionResult{Supported: true, Error: "extension name is required"}
	}

	commandName := "phpenmod"
	verb := "enabled"
	if !request.Enabled {
		commandName = "phpdismod"
		verb = "disabled"
	}

	if _, err := p.runner.LookPath(commandName); err != nil {
		return domainphpmanager.ActionResult{Supported: false, Message: fmt.Sprintf("%s is not available on this Linux distribution", commandName)}
	}

	installedVersions, discoverErr := p.discoverInstalledVersions(ctx)
	if discoverErr != nil {
		installedVersions = []string{}
	}
	if len(installedVersions) == 0 {
		versionOutput, versionErr := p.runner.Run(ctx, "php", "-v")
		if versionErr == nil {
			activeVersion := parsePHPVersionFromOutput(versionOutput)
			if activeVersion != "" {
				installedVersions = []string{activeVersion}
			}
		}
	}

	commands := make([]string, 0, len(installedVersions))
	usedPkexecForExtensions := false
	for _, version := range installedVersions {
		args := []string{"-v", version, "-s", "ALL", extensionName}
		commands = append(commands, commandName+" "+strings.Join(args, " "))
		usedPkexec, err := p.runWithPrivilegeFallback(ctx, commandName, args...)
		if usedPkexec {
			usedPkexecForExtensions = true
		}
		if err != nil {
			result := domainphpmanager.ActionResult{Supported: true, Command: commandName + " " + strings.Join(args, " "), Error: fmt.Sprintf("failed to update extension %s for PHP %s: %v", extensionName, version, err)}
			if requiresRootPrivileges(err) {
				result.RequiresPrivilege = true
				result.SuggestedCommands = []string{"sudo " + commandName + " " + strings.Join(args, " ")}
			}
			return result
		}
	}

	services := discoverFPMServiceTargets()
	usedPkexecForRestart := false
	for _, service := range services {
		usedPkexec, restartErr := p.restartFPMService(ctx, service.ServiceName)
		if usedPkexec {
			usedPkexecForRestart = true
		}
		if restartErr != nil {
			return domainphpmanager.ActionResult{
				Supported:         true,
				RequiresPrivilege: true,
				SuggestedCommands: []string{service.RestartCommand},
				Error:             fmt.Sprintf("extension updated, but failed to restart %s: %v", service.ServiceName, restartErr),
				Message:           "extension updated, but one or more PHP-FPM services require manual restart",
			}
		}
	}
	message := fmt.Sprintf("extension %s %s successfully", extensionName, verb)
	if usedPkexecForExtensions || usedPkexecForRestart {
		message += " (via pkexec)"
	}

	return domainphpmanager.ActionResult{Success: true, Supported: true, Command: strings.Join(commands, " && "), Message: message}
}

func (p *Provider) discoverInstalledVersions(ctx context.Context) ([]string, error) {
	output, err := p.runner.Run(ctx, "dpkg-query", "-W", "-f=${Package}\\n")
	if err != nil {
		return nil, fmt.Errorf("dpkg-query failed: %w", err)
	}
	return uniqueSortedVersions(parseVersionsFromPackageList(output)), nil
}

func (p *Provider) discoverAvailableVersions(ctx context.Context) ([]string, error) {
	output, err := p.runner.Run(ctx, "apt-cache", "search", "--names-only", "^php[0-9]\\.[0-9]-cli$")
	if err != nil {
		return nil, fmt.Errorf("apt-cache search failed: %w", err)
	}
	return uniqueSortedVersions(parseVersionsFromPackageList(output)), nil
}

func (p *Provider) isAptBasedLinux(ctx context.Context) bool {
	if _, err := p.runner.LookPath("apt-get"); err != nil {
		return false
	}
	if _, err := p.runner.Run(ctx, "dpkg-query", "--version"); err != nil {
		return false
	}
	return true
}

func (p *Provider) isVersionInstalled(ctx context.Context, version string) (bool, error) {
	packageName := fmt.Sprintf("php%s-cli", version)
	_, err := p.runner.Run(ctx, "dpkg-query", "-W", "-f=${Status}", packageName)
	if err != nil {
		return false, err
	}
	return true, nil
}

func parsePHPVersionFromOutput(output string) string {
	line := firstLine(output)
	if line == "" {
		return ""
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}
	version := strings.TrimSpace(fields[1])
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

func parseVersionsFromPackageList(output string) []string {
	versions := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		name := trimmed
		if idx := strings.Index(trimmed, " "); idx > 0 {
			name = trimmed[:idx]
		}
		if !strings.HasPrefix(name, "php") || !strings.HasSuffix(name, "-cli") {
			continue
		}
		version := strings.TrimSuffix(strings.TrimPrefix(name, "php"), "-cli")
		if isValidVersion(version) {
			versions = append(versions, version)
		}
	}
	return versions
}

type fpmServiceTarget struct {
	ConfDPath      string
	ServiceName    string
	RestartCommand string
}

func discoverFPMServiceTargets() []fpmServiceTarget {
	matches, err := filepath.Glob("/etc/php/*/fpm/conf.d")
	if err != nil {
		return nil
	}

	targets := make([]fpmServiceTarget, 0, len(matches))
	for _, confD := range matches {
		version := filepath.Base(filepath.Dir(filepath.Dir(confD)))
		if !isValidVersion(version) {
			continue
		}
		serviceName := fmt.Sprintf("php%s-fpm", version)
		targets = append(targets, fpmServiceTarget{
			ConfDPath:      confD,
			ServiceName:    serviceName,
			RestartCommand: "sudo systemctl restart " + serviceName,
		})
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].ServiceName < targets[j].ServiceName
	})

	return targets
}

func uniqueSortedVersions(versions []string) []string {
	seen := make(map[string]struct{}, len(versions))
	unique := make([]string, 0, len(versions))
	for _, version := range versions {
		if _, ok := seen[version]; ok {
			continue
		}
		seen[version] = struct{}{}
		unique = append(unique, version)
	}
	sort.Slice(unique, func(i, j int) bool {
		return compareVersions(unique[i], unique[j]) > 0
	})
	return unique
}

func compareVersions(left string, right string) int {
	leftMajor, leftMinor := parseVersionParts(left)
	rightMajor, rightMinor := parseVersionParts(right)
	if leftMajor != rightMajor {
		if leftMajor > rightMajor {
			return 1
		}
		return -1
	}
	if leftMinor != rightMinor {
		if leftMinor > rightMinor {
			return 1
		}
		return -1
	}
	return 0
}

func parseVersionParts(version string) (int, int) {
	parts := strings.Split(version, ".")
	if len(parts) != 2 {
		return 0, 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		major = 0
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		minor = 0
	}
	return major, minor
}

func isValidVersion(version string) bool {
	return versionPattern.MatchString(version)
}

func phpSettingsReadScript() string {
	return `echo "upload_max_filesize=".ini_get("upload_max_filesize").PHP_EOL;` +
		`echo "post_max_size=".ini_get("post_max_size").PHP_EOL;` +
		`echo "memory_limit=".ini_get("memory_limit").PHP_EOL;` +
		`echo "max_execution_time=".ini_get("max_execution_time").PHP_EOL;`
}

func parsePHPIniSettingsOutput(output string) domainphpmanager.IniSettings {
	settings := domainphpmanager.IniSettings{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "upload_max_filesize":
			settings.UploadMaxFilesize = value
		case "post_max_size":
			settings.PostMaxSize = value
		case "memory_limit":
			settings.MemoryLimit = value
		case "max_execution_time":
			settings.MaxExecutionTime = value
		}
	}

	return settings
}

func parsePHPExtensionsOutput(output string) map[string]struct{} {
	enabled := make(map[string]struct{})
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			continue
		}
		name := normalizeExtensionName(trimmed)
		if name == "" {
			continue
		}
		enabled[name] = struct{}{}
	}

	return enabled
}

func discoverAvailableExtensionINIFiles() (map[string]string, error) {
	files, err := filepath.Glob("/etc/php/*/mods-available/*.ini")
	if err != nil {
		return nil, fmt.Errorf("failed to discover extension ini files: %w", err)
	}

	available := make(map[string]string, len(files))
	for _, file := range files {
		name := normalizeExtensionName(strings.TrimSuffix(filepath.Base(file), ".ini"))
		if name == "" {
			continue
		}
		if _, exists := available[name]; exists {
			continue
		}
		available[name] = file
	}

	return available, nil
}

func normalizeExtensionName(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	trimmed = strings.TrimSuffix(trimmed, ".ini")
	return trimmed
}

func buildManagedPHPSettingsINI(request domainphpmanager.IniSettingsUpdateRequest) (string, error) {
	settings := map[string]string{
		"upload_max_filesize": strings.TrimSpace(request.UploadMaxFilesize),
		"post_max_size":       strings.TrimSpace(request.PostMaxSize),
		"memory_limit":        strings.TrimSpace(request.MemoryLimit),
		"max_execution_time":  strings.TrimSpace(request.MaxExecutionTime),
	}

	if settings["upload_max_filesize"] == "" && settings["post_max_size"] == "" && settings["memory_limit"] == "" && settings["max_execution_time"] == "" {
		return "", fmt.Errorf("at least one php.ini setting is required")
	}

	orderedKeys := []string{"upload_max_filesize", "post_max_size", "memory_limit", "max_execution_time"}
	lines := []string{"; Managed by Phant"}
	for _, key := range orderedKeys {
		if settings[key] == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s = %s", key, settings[key]))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func (p *Provider) writeManagedINI(ctx context.Context, targetPath string, content string) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err == nil {
		return false, nil
	} else {
		if !isPermissionError(err) {
			return false, err
		}

		if !p.canUsePkexec() {
			return false, err
		}

		tmpFile, tmpErr := os.CreateTemp("", "phant-php-settings-*.ini")
		if tmpErr != nil {
			return false, tmpErr
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, writeErr := tmpFile.WriteString(content); writeErr != nil {
			tmpFile.Close()
			return false, writeErr
		}
		if closeErr := tmpFile.Close(); closeErr != nil {
			return false, closeErr
		}

		if _, mkdirErr := p.runner.Run(ctx, "pkexec", "mkdir", "-p", filepath.Dir(targetPath)); mkdirErr != nil {
			return false, fmt.Errorf("pkexec mkdir failed: %w", mkdirErr)
		}
		if _, installErr := p.runner.Run(ctx, "pkexec", "install", "-m", "0644", tmpPath, targetPath); installErr != nil {
			return false, fmt.Errorf("pkexec install failed: %w", installErr)
		}

		return true, nil
	}
}

func parseAdditionalINIPath(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		const prefix = "Scan for additional .ini files in:"
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	return ""
}

func buildLinuxWriteManagedINICommand(targetPath string, content string) string {
	escapedDir := shellSingleQuote(filepath.Dir(targetPath))
	escapedTarget := shellSingleQuote(targetPath)
	escapedContent := shellSingleQuote(content)

	return fmt.Sprintf("sudo mkdir -p %s && printf '%%s' %s | sudo tee %s > /dev/null", escapedDir, escapedContent, escapedTarget)
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	if os.IsPermission(err) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "operation not permitted") ||
		strings.Contains(lower, "access is denied")
}

func (p *Provider) restartFPMService(ctx context.Context, serviceName string) (bool, error) {
	if _, err := p.runner.LookPath("systemctl"); err != nil {
		return false, nil
	}
	_, err := p.runner.Run(ctx, "systemctl", "restart", serviceName)
	if err == nil {
		return false, nil
	}

	if !requiresRootPrivileges(err) || !p.canUsePkexec() {
		return false, err
	}

	_, pkexecErr := p.runner.Run(ctx, "pkexec", "systemctl", "restart", serviceName)
	if pkexecErr != nil {
		return false, fmt.Errorf("%v; pkexec restart failed: %w", err, pkexecErr)
	}

	return true, nil
}

func (p *Provider) runWithPrivilegeFallback(ctx context.Context, name string, args ...string) (bool, error) {
	_, err := p.runner.Run(ctx, name, args...)
	if err == nil {
		return false, nil
	}

	if !requiresRootPrivileges(err) || !p.canUsePkexec() {
		return false, err
	}

	pkexecArgs := make([]string, 0, len(args)+1)
	pkexecArgs = append(pkexecArgs, name)
	pkexecArgs = append(pkexecArgs, args...)
	_, pkexecErr := p.runner.Run(ctx, "pkexec", pkexecArgs...)
	if pkexecErr != nil {
		return false, fmt.Errorf("%v; pkexec failed: %w", err, pkexecErr)
	}

	return true, nil
}

func (p *Provider) canUsePkexec() bool {
	_, err := p.runner.LookPath("pkexec")
	return err == nil
}

func firstLine(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	return strings.TrimSpace(lines[0])
}

func requiresRootPrivileges(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "permission denied") ||
		strings.Contains(lower, "are you root") ||
		strings.Contains(lower, "could not open lock file") ||
		strings.Contains(lower, "superuser")
}

func unsupportedActionResult(action string, version string) domainphpmanager.ActionResult {
	return domainphpmanager.ActionResult{
		Supported: false,
		Version:   version,
		Message:   fmt.Sprintf("PHP manager action %q is currently supported on Linux only.", action),
	}
}

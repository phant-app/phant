package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type ValetLinuxVerification struct {
	GeneratedAt         string             `json:"generatedAt"`
	Supported           bool               `json:"supported"`
	ValetDetected       bool               `json:"valetDetected"`
	ServiceManager      string             `json:"serviceManager"`
	CLIConfDPath        string             `json:"cliConfDPath"`
	CLIAutoPrepend      string             `json:"cliAutoPrepend"`
	ExpectedPrependPath string             `json:"expectedPrependPath"`
	FPMServices         []FPMServiceStatus `json:"fpmServices"`
	Recommendations     []string           `json:"recommendations"`
	LastError           string             `json:"lastError"`
}

type FPMServiceStatus struct {
	ServiceName         string `json:"serviceName"`
	Version             string `json:"version"`
	ConfDPath           string `json:"confDPath"`
	HookIniPath         string `json:"hookIniPath"`
	HookIniExists       bool   `json:"hookIniExists"`
	AutoPrependFile     string `json:"autoPrependFile"`
	MatchesExpected     bool   `json:"matchesExpected"`
	SystemdActive       bool   `json:"systemdActive"`
	SystemdEnabled      bool   `json:"systemdEnabled"`
	RestartCommand      string `json:"restartCommand"`
	VerificationCommand string `json:"verificationCommand"`
}

type ValetLinuxRemediationResult struct {
	GeneratedAt         string                   `json:"generatedAt"`
	Supported           bool                     `json:"supported"`
	Confirmed           bool                     `json:"confirmed"`
	Applied             bool                     `json:"applied"`
	ExpectedPrependPath string                   `json:"expectedPrependPath"`
	RequiresSudo        bool                     `json:"requiresSudo"`
	SuggestedCommands   []string                 `json:"suggestedCommands"`
	Targets             []ValetRemediationTarget `json:"targets"`
	Message             string                   `json:"message"`
	Error               string                   `json:"error"`
}

type ValetRemediationTarget struct {
	ServiceName      string `json:"serviceName"`
	HookIniPath      string `json:"hookIniPath"`
	WriteAttempted   bool   `json:"writeAttempted"`
	Written          bool   `json:"written"`
	WriteError       string `json:"writeError"`
	RestartAttempted bool   `json:"restartAttempted"`
	Restarted        bool   `json:"restarted"`
	RestartError     string `json:"restartError"`
	RestartCommand   string `json:"restartCommand"`
}

func VerifyValetLinux(ctx context.Context) ValetLinuxVerification {
	report := ValetLinuxVerification{
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		ServiceManager: detectServiceManager(ctx),
		Supported:      runtime.GOOS == "linux",
	}

	if runtime.GOOS != "linux" {
		report.Recommendations = append(report.Recommendations, "Valet Linux verification is available on Linux only.")
		return report
	}

	report.ValetDetected = isCommandAvailable("valet")

	configDir, err := os.UserConfigDir()
	if err != nil {
		report.LastError = fmt.Sprintf("failed to resolve user config dir: %v", err)
		return report
	}
	report.ExpectedPrependPath = filepath.Join(configDir, "phant", "php", "phant_prepend.php")

	cliIniOutput, iniErr := runCommand(ctx, "php", "--ini")
	if iniErr != nil {
		report.LastError = fmt.Sprintf("php --ini failed: %v", iniErr)
	} else {
		report.CLIConfDPath = parseAdditionalINIPath(cliIniOutput)
		report.CLIAutoPrepend = readAutoPrependFromIni(filepath.Join(report.CLIConfDPath, "99-phant.ini"))
	}

	services, discoverErr := discoverFPMServices(ctx)
	if discoverErr != nil && report.LastError == "" {
		report.LastError = discoverErr.Error()
	}
	report.FPMServices = services

	if !report.ValetDetected {
		report.Recommendations = append(report.Recommendations, "Valet CLI was not found in PATH. Install or expose Valet Linux to verify HTTP capture via Valet.")
	}

	if len(report.FPMServices) == 0 {
		report.Recommendations = append(report.Recommendations, "No PHP-FPM installations were discovered under /etc/php/*/fpm. If Valet uses a custom PHP runtime, verify its conf.d path manually.")
	} else {
		for _, service := range report.FPMServices {
			if !service.HookIniExists || !service.MatchesExpected {
				report.Recommendations = append(report.Recommendations,
					fmt.Sprintf("%s is not wired to Phant prepend script. Ensure %s sets auto_prepend_file to %s.", service.ServiceName, service.HookIniPath, report.ExpectedPrependPath),
				)
			}
			if !service.SystemdActive {
				report.Recommendations = append(report.Recommendations,
					fmt.Sprintf("%s is not active. Start/restart with: %s", service.ServiceName, service.RestartCommand),
				)
			}
		}
	}

	if report.LastError == "" && len(report.Recommendations) == 0 {
		report.Recommendations = append(report.Recommendations, "Valet Linux verification looks healthy. Run an HTTP route with dump()/dd() to confirm end-to-end capture.")
	}

	return report
}

func ApplyValetLinuxRemediation(ctx context.Context, confirm bool) ValetLinuxRemediationResult {
	result := ValetLinuxRemediationResult{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Supported:   runtime.GOOS == "linux",
		Confirmed:   confirm,
	}

	if runtime.GOOS != "linux" {
		result.Message = "Valet remediation is supported on Linux only."
		return result
	}

	result.ExpectedPrependPath = expectedPrependPath()
	if result.ExpectedPrependPath == "" {
		result.Error = "unable to resolve expected prepend path"
		return result
	}

	if !confirm {
		result.Message = "Confirmation required before applying Valet Linux remediation."
		return result
	}

	services, err := discoverFPMServices(ctx)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if len(services) == 0 {
		result.Message = "No PHP-FPM services were discovered under /etc/php/*/fpm."
		return result
	}

	desired := buildConfDContent(result.ExpectedPrependPath)
	allSucceeded := true
	changesApplied := false

	for _, service := range services {
		target := ValetRemediationTarget{
			ServiceName:    service.ServiceName,
			HookIniPath:    service.HookIniPath,
			RestartCommand: service.RestartCommand,
		}

		needsWrite := !service.HookIniExists || !service.MatchesExpected
		if needsWrite {
			target.WriteAttempted = true

			writeErr := writeHookINI(ctx, service.HookIniPath, desired)
			if writeErr != nil {
				target.WriteError = writeErr.Error()
				allSucceeded = false

				if isPermissionError(writeErr) {
					result.RequiresSudo = true
					result.SuggestedCommands = append(result.SuggestedCommands, buildLinuxManualCommand(service.HookIniPath, result.ExpectedPrependPath))
				}
			} else {
				target.Written = true
				changesApplied = true
			}
		}

		shouldRestart := target.Written
		if shouldRestart {
			target.RestartAttempted = true

			if restartErr := restartFPMService(ctx, service.ServiceName); restartErr != nil {
				target.RestartError = restartErr.Error()
				allSucceeded = false

				if strings.Contains(strings.ToLower(restartErr.Error()), "permission") || strings.Contains(strings.ToLower(restartErr.Error()), "authentication") {
					result.RequiresSudo = true
					result.SuggestedCommands = append(result.SuggestedCommands, service.RestartCommand)
				}
			} else {
				target.Restarted = true
			}
		}

		result.Targets = append(result.Targets, target)
	}

	if allSucceeded {
		result.Applied = true
		if changesApplied {
			result.Message = "Valet Linux remediation applied successfully."
		} else {
			result.Message = "No changes required. PHP-FPM hooks are already configured."
		}
		return result
	}

	if result.Message == "" {
		result.Message = "Valet Linux remediation completed with some failures. Review target errors and suggested commands."
	}

	result.SuggestedCommands = uniqueStrings(result.SuggestedCommands)
	return result
}

func discoverFPMServices(ctx context.Context) ([]FPMServiceStatus, error) {
	versionDirs, err := filepath.Glob("/etc/php/*/fpm")
	if err != nil {
		return nil, fmt.Errorf("failed to inspect /etc/php for fpm: %w", err)
	}

	services := make([]FPMServiceStatus, 0, len(versionDirs))
	for _, dir := range versionDirs {
		version := filepath.Base(filepath.Dir(dir))
		serviceName := fmt.Sprintf("php%s-fpm.service", version)
		hookIniPath := filepath.Join(dir, "conf.d", "99-phant.ini")
		autoPrepend := readAutoPrependFromIni(hookIniPath)
		active := isSystemdUnitActive(ctx, serviceName)
		enabled := isSystemdUnitEnabled(ctx, serviceName)
		services = append(services, FPMServiceStatus{
			ServiceName:         serviceName,
			Version:             version,
			ConfDPath:           filepath.Join(dir, "conf.d"),
			HookIniPath:         hookIniPath,
			HookIniExists:       fileExists(hookIniPath),
			AutoPrependFile:     autoPrepend,
			MatchesExpected:     autoPrepend != "" && autoPrepend == expectedPrependPath(),
			SystemdActive:       active,
			SystemdEnabled:      enabled,
			RestartCommand:      fmt.Sprintf("sudo systemctl restart %s", serviceName),
			VerificationCommand: fmt.Sprintf("php-fpm%s -i | grep auto_prepend_file", version),
		})
	}

	preferredVersion := strings.TrimSpace(detectDefaultPHPVersion(ctx))
	sort.Slice(services, func(i, j int) bool {
		leftPreferred := preferredVersion != "" && services[i].Version == preferredVersion
		rightPreferred := preferredVersion != "" && services[j].Version == preferredVersion
		if leftPreferred != rightPreferred {
			return leftPreferred
		}

		return services[i].Version < services[j].Version
	})

	return services, nil
}

func expectedPrependPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "phant", "php", "phant_prepend.php")
}

func readAutoPrependFromIni(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(trimmed), "auto_prepend_file") {
			continue
		}

		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		return value
	}

	return ""
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func isSystemdUnitActive(ctx context.Context, unit string) bool {
	if _, err := runCommand(ctx, "systemctl", "is-active", unit); err != nil {
		return false
	}
	return true
}

func isSystemdUnitEnabled(ctx context.Context, unit string) bool {
	if _, err := runCommand(ctx, "systemctl", "is-enabled", unit); err != nil {
		return false
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeHookINI(ctx context.Context, targetPath string, content string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		if isPermissionError(err) {
			return privilegedWriteHookINI(ctx, targetPath, content)
		}
		return err
	}

	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		if isPermissionError(err) {
			return privilegedWriteHookINI(ctx, targetPath, content)
		}
		return err
	}

	return nil
}

func privilegedWriteHookINI(ctx context.Context, targetPath string, content string) error {
	pkexecPath, lookErr := exec.LookPath("pkexec")
	if lookErr != nil {
		return fmt.Errorf("pkexec not found for privileged write")
	}

	tmpFile, tmpErr := os.CreateTemp("", "phant-fpm-hook-*.ini")
	if tmpErr != nil {
		return tmpErr
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, writeErr := tmpFile.WriteString(content); writeErr != nil {
		tmpFile.Close()
		return writeErr
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		return closeErr
	}

	if _, mkdirErr := runCommand(ctx, pkexecPath, "mkdir", "-p", filepath.Dir(targetPath)); mkdirErr != nil {
		return mkdirErr
	}

	if _, installErr := runCommand(ctx, pkexecPath, "install", "-m", "0644", tmpPath, targetPath); installErr != nil {
		return installErr
	}

	return nil
}

func restartFPMService(ctx context.Context, serviceName string) error {
	if _, err := runCommand(ctx, "systemctl", "restart", serviceName); err == nil {
		return nil
	}

	pkexecPath, lookErr := exec.LookPath("pkexec")
	if lookErr != nil {
		return fmt.Errorf("systemctl restart failed and pkexec not found")
	}

	_, err := runCommand(ctx, pkexecPath, "systemctl", "restart", serviceName)
	return err
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}

	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}

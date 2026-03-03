package setup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	phantMarkerBegin = "; BEGIN PHANT AUTO_PREPEND"
	phantMarkerEnd   = "; END PHANT AUTO_PREPEND"
)

type HookInstallResult struct {
	Success            bool   `json:"success"`
	AlreadyEnabled     bool   `json:"alreadyEnabled"`
	PHPIniPath         string `json:"phpIniPath"`
	PrependPath        string `json:"prependPath"`
	BackupPath         string `json:"backupPath"`
	SocketPath         string `json:"socketPath"`
	RequiresSudo       bool   `json:"requiresSudo"`
	SuggestedCmd       string `json:"suggestedCmd"`
	PrivilegeStrategy  string `json:"privilegeStrategy"`
	PrivilegeAttempted bool   `json:"privilegeAttempted"`
	Message            string `json:"message"`
	Error              string `json:"error"`
}

func InstallCLIHook(ctx context.Context, socketPath string) HookInstallResult {
	result := HookInstallResult{
		SocketPath:        socketPath,
		PrivilegeStrategy: currentPrivilegeStrategy(),
	}

	iniOutput, err := runCommand(ctx, "php", "--ini")
	if err != nil {
		result.Error = fmt.Sprintf("php --ini failed: %v", err)
		return result
	}

	phpIniPath := parseLoadedConfigurationFile(iniOutput)
	if phpIniPath == "" || strings.EqualFold(phpIniPath, "(none)") {
		result.Error = "unable to detect loaded php.ini file"
		return result
	}

	prependPath, writeErr := writePrependScript(socketPath)
	if writeErr != nil {
		result.Error = writeErr.Error()
		return result
	}
	result.PrependPath = prependPath

	scanDir := parseAdditionalINIPath(iniOutput)
	if scanDir != "" && !strings.EqualFold(scanDir, "(none)") {
		result.PHPIniPath = filepath.Join(scanDir, "99-phant.ini")
		return installInConfD(ctx, result, prependPath)
	}

	result.PHPIniPath = phpIniPath
	return installByPatchingPHPIni(result, prependPath)
}

func installInConfD(ctx context.Context, result HookInstallResult, prependPath string) HookInstallResult {
	desired := buildConfDContent(prependPath)

	originalContent, readErr := os.ReadFile(result.PHPIniPath)
	if readErr == nil {
		if string(originalContent) == desired {
			result.Success = true
			result.AlreadyEnabled = true
			result.Message = "CLI hook already enabled"
			return result
		}

		backupPath := fmt.Sprintf("%s.phant.bak.%s", result.PHPIniPath, time.Now().UTC().Format("20060102150405"))
		if writeErr := os.WriteFile(backupPath, originalContent, 0o644); writeErr != nil {
			if isPermissionError(writeErr) {
				result = tryPrivilegedConfDInstall(ctx, result, desired)
				if result.Success {
					return result
				}
			}
			result.Error = formatPermissionAwareError("failed to create conf.d backup", writeErr)
			result.RequiresSudo = true
			result.SuggestedCmd = buildManualCommandForOS(result.PHPIniPath, prependPath)
			return result
		}
		result.BackupPath = backupPath
	} else if !errors.Is(readErr, os.ErrNotExist) {
		if isPermissionError(readErr) {
			result = tryPrivilegedConfDInstall(ctx, result, desired)
			if result.Success {
				return result
			}
			result.Error = formatPermissionAwareError("failed to read conf.d file", readErr)
			result.RequiresSudo = true
			result.SuggestedCmd = buildManualCommandForOS(result.PHPIniPath, prependPath)
			return result
		}
		result.Error = fmt.Sprintf("failed to read conf.d file: %v", readErr)
		return result
	}

	if writeErr := os.WriteFile(result.PHPIniPath, []byte(desired), 0o644); writeErr != nil {
		if isPermissionError(writeErr) {
			result = tryPrivilegedConfDInstall(ctx, result, desired)
			if result.Success {
				return result
			}
			result.RequiresSudo = true
			result.SuggestedCmd = buildManualCommandForOS(result.PHPIniPath, prependPath)
		}
		result.Error = formatPermissionAwareError("failed to write conf.d file", writeErr)
		return result
	}

	result.Success = true
	result.Message = "CLI hook enabled via conf.d file. New PHP CLI processes should emit dump events to Phant."
	return result
}

func tryPrivilegedConfDInstall(ctx context.Context, result HookInstallResult, desired string) HookInstallResult {
	result.PrivilegeAttempted = true

	switch runtime.GOOS {
	case "linux":
		pkexecPath, lookErr := exec.LookPath("pkexec")
		if lookErr != nil {
			result.Error = "pkexec not found on this Linux system"
			return result
		}

		tmpFile, tmpErr := os.CreateTemp("", "phant-confd-*.ini")
		if tmpErr != nil {
			result.Error = fmt.Sprintf("failed to create temp config file: %v", tmpErr)
			return result
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, writeErr := tmpFile.WriteString(desired); writeErr != nil {
			tmpFile.Close()
			result.Error = fmt.Sprintf("failed to write temp config file: %v", writeErr)
			return result
		}
		if closeErr := tmpFile.Close(); closeErr != nil {
			result.Error = fmt.Sprintf("failed to close temp config file: %v", closeErr)
			return result
		}

		if _, mkdirErr := runCommand(ctx, pkexecPath, "mkdir", "-p", filepath.Dir(result.PHPIniPath)); mkdirErr != nil {
			result.Error = fmt.Sprintf("pkexec mkdir failed: %v", mkdirErr)
			return result
		}

		if _, installErr := runCommand(ctx, pkexecPath, "install", "-m", "0644", tmpPath, result.PHPIniPath); installErr != nil {
			result.Error = fmt.Sprintf("pkexec install failed: %v", installErr)
			return result
		}

		result.Success = true
		result.RequiresSudo = false
		result.Message = "CLI hook enabled via conf.d file using pkexec. New PHP CLI processes should emit dump events to Phant."
		return result
	case "darwin":
		result.Error = "automatic privileged install is not implemented on macOS yet"
		return result
	case "windows":
		result.Error = "automatic privileged install is not implemented on Windows yet"
		return result
	default:
		result.Error = fmt.Sprintf("automatic privileged install is not implemented for OS: %s", runtime.GOOS)
		return result
	}
}

func installByPatchingPHPIni(result HookInstallResult, prependPath string) HookInstallResult {
	originalContent, readErr := os.ReadFile(result.PHPIniPath)
	if readErr != nil {
		result.Error = fmt.Sprintf("failed to read php.ini: %v", readErr)
		return result
	}

	patchedContent, changed := ensureAutoPrependBlock(string(originalContent), prependPath)
	if !changed {
		result.Success = true
		result.AlreadyEnabled = true
		result.Message = "CLI hook already enabled"
		return result
	}

	backupPath := fmt.Sprintf("%s.phant.bak.%s", result.PHPIniPath, time.Now().UTC().Format("20060102150405"))
	if writeErr := os.WriteFile(backupPath, originalContent, 0o644); writeErr != nil {
		if isPermissionError(writeErr) {
			result.RequiresSudo = true
			result.SuggestedCmd = buildManualCommandForOS(result.PHPIniPath, prependPath)
		}
		result.Error = formatPermissionAwareError("failed to create php.ini backup", writeErr)
		return result
	}
	result.BackupPath = backupPath

	if writeErr := os.WriteFile(result.PHPIniPath, []byte(patchedContent), 0o644); writeErr != nil {
		if isPermissionError(writeErr) {
			result.RequiresSudo = true
			result.SuggestedCmd = buildManualCommandForOS(result.PHPIniPath, prependPath)
		}
		_ = os.WriteFile(result.PHPIniPath, originalContent, 0o644)
		result.Error = formatPermissionAwareError("failed to patch php.ini", writeErr)
		return result
	}

	result.Success = true
	result.Message = "CLI hook enabled by patching php.ini. New PHP processes should emit dump events to Phant."
	return result
}

func buildConfDContent(prependPath string) string {
	return fmt.Sprintf("; Generated by Phant\nauto_prepend_file = \"%s\"\n", prependPath)
}

func formatPermissionAwareError(prefix string, err error) string {
	if isPermissionError(err) {
		return fmt.Sprintf("%s: %v (sudo/root permission required)", prefix, err)
	}
	return fmt.Sprintf("%s: %v", prefix, err)
}

func isPermissionError(err error) bool {
	return errors.Is(err, os.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "permission denied")
}

func currentPrivilegeStrategy() string {
	switch runtime.GOOS {
	case "linux":
		return "linux:pkexec"
	case "darwin":
		return "darwin:manual"
	case "windows":
		return "windows:manual"
	default:
		return runtime.GOOS + ":manual"
	}
}

func buildManualCommandForOS(confDPath string, prependPath string) string {
	switch runtime.GOOS {
	case "linux":
		return buildLinuxManualCommand(confDPath, prependPath)
	case "darwin":
		return fmt.Sprintf("sudo sh -c 'printf \"%%s\\n\" \"; Generated by Phant\" \"auto_prepend_file = \\\"%s\\\"\" > \"%s\"'", prependPath, confDPath)
	case "windows":
		return "Run Phant as Administrator and enable the CLI hook again."
	default:
		return "Enable privileged write manually for your OS and create the conf.d entry shown in docs."
	}
}

func buildLinuxManualCommand(confDPath string, prependPath string) string {
	escapedConfDPath := shellSingleQuote(confDPath)
	confDDir := shellSingleQuote(filepath.Dir(confDPath))
	entry := shellSingleQuote(fmt.Sprintf("auto_prepend_file = \"%s\"", prependPath))

	return fmt.Sprintf(
		"sudo mkdir -p %s && printf '%%s\\n' '; Generated by Phant' %s | sudo tee %s > /dev/null",
		confDDir,
		entry,
		escapedConfDPath,
	)
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func writePrependScript(socketPath string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}

	targetDir := filepath.Join(configDir, "phant", "php")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create prepend directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, "phant_prepend.php")
	content := strings.ReplaceAll(phpPrependTemplate, "{{SOCKET_PATH}}", socketPath)
	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write prepend script: %w", err)
	}

	return targetPath, nil
}

func parseLoadedConfigurationFile(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Loaded Configuration File:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "Loaded Configuration File:"))
		}
	}
	return ""
}

func parseAdditionalINIPath(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Scan for additional .ini files in:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "Scan for additional .ini files in:"))
		}
	}
	return ""
}

func ensureAutoPrependBlock(content string, prependPath string) (string, bool) {
	block := fmt.Sprintf("%s\nauto_prepend_file = \"%s\"\n%s", phantMarkerBegin, prependPath, phantMarkerEnd)

	startIdx := strings.Index(content, phantMarkerBegin)
	endIdx := strings.Index(content, phantMarkerEnd)
	if startIdx >= 0 && endIdx > startIdx {
		replaceEnd := endIdx + len(phantMarkerEnd)
		next := content[:startIdx] + block + content[replaceEnd:]
		return next, next != content
	}

	trimmed := strings.TrimRight(content, "\n")
	next := trimmed + "\n\n" + block + "\n"
	return next, next != content
}

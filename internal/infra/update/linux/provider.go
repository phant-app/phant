package linux

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	domainupdate "phant/internal/domain/update"
	"phant/internal/infra/system"
)

type Provider struct {
	runner system.Runner
	client *http.Client
}

func NewProvider(runner system.Runner) *Provider {
	return &Provider{
		runner: runner,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Platform() string {
	return p.runner.GOOS()
}

func (p *Provider) HTTPClient() *http.Client {
	return p.client
}

func (p *Provider) InstallDownloaded(ctx context.Context, downloadedPath string) domainupdate.InstallResult {
	candidate := strings.TrimSpace(downloadedPath)
	if candidate == "" {
		return domainupdate.InstallResult{Error: "downloaded update file path is required"}
	}
	if p.Platform() != "linux" {
		return domainupdate.InstallResult{Error: "install update is currently supported only on Linux"}
	}

	sourcePath, err := filepath.Abs(candidate)
	if err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("resolve downloaded file path: %v", err)}
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("read downloaded file: %v", err)}
	}
	if sourceInfo.IsDir() {
		return domainupdate.InstallResult{Error: "downloaded update path must be a file"}
	}

	currentExecutable, err := os.Executable()
	if err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("resolve current executable path: %v", err)}
	}
	currentPath, err := filepath.Abs(currentExecutable)
	if err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("resolve current executable path: %v", err)}
	}
	if sourcePath == currentPath {
		return domainupdate.InstallResult{Error: "downloaded file matches current executable path"}
	}

	installerScript, err := os.CreateTemp("", "phant-install-update-*.sh")
	if err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("create installer script: %v", err)}
	}

	script := buildInstallScript(sourcePath, currentPath)
	if _, err := installerScript.WriteString(script); err != nil {
		_ = installerScript.Close()
		return domainupdate.InstallResult{Error: fmt.Sprintf("write installer script: %v", err)}
	}
	if err := installerScript.Close(); err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("close installer script: %v", err)}
	}
	if err := os.Chmod(installerScript.Name(), 0o700); err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("set installer script permissions: %v", err)}
	}

	if _, err := p.runner.Run(ctx, "nohup", "sh", installerScript.Name()); err != nil {
		return domainupdate.InstallResult{Error: fmt.Sprintf("launch installer script: %v", err)}
	}

	return domainupdate.InstallResult{
		Installed:  true,
		TargetPath: currentPath,
		Message:    "Update install started. Phant will restart automatically.",
	}
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func buildInstallScript(sourcePath string, targetPath string) string {
	quotedSource := shellSingleQuote(sourcePath)
	quotedTarget := shellSingleQuote(targetPath)
	quotedBackup := shellSingleQuote(targetPath + ".bak")
	quotedTemp := shellSingleQuote(targetPath + ".new")

	return "#!/bin/sh\n" +
		"set -eu\n" +
		"sleep 2\n" +
		"if [ ! -f " + quotedSource + " ]; then\n" +
		"  exit 1\n" +
		"fi\n" +
		"cp " + quotedSource + " " + quotedTemp + "\n" +
		"chmod 755 " + quotedTemp + "\n" +
		"if [ -f " + quotedTarget + " ]; then\n" +
		"  cp " + quotedTarget + " " + quotedBackup + "\n" +
		"fi\n" +
		"mv " + quotedTemp + " " + quotedTarget + "\n" +
		"nohup " + quotedTarget + " >/dev/null 2>&1 &\n"
}

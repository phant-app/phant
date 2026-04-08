package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	domainlicense "phant/internal/domain/license"
	domainupdate "phant/internal/domain/update"
)

const DefaultManifestURL = "https://phant.app/update.json"

type Dependencies struct {
	CurrentVersion func() string
	GetLicenseKey  func(context.Context) domainlicense.KeyResult
	HTTPClient     func() *http.Client
}

type Service struct {
	deps Dependencies
}

func NewService(deps Dependencies) *Service {
	if deps.CurrentVersion == nil {
		deps.CurrentVersion = func() string { return "" }
	}
	if deps.HTTPClient == nil {
		deps.HTTPClient = func() *http.Client { return &http.Client{} }
	}
	return &Service{deps: deps}
}

func (s *Service) CheckForUpdate(ctx context.Context, manifestURL string) domainupdate.CheckResult {
	manifest, err := s.fetchManifest(ctx, manifestURL)
	if err != nil {
		return domainupdate.CheckResult{
			CurrentVersion: normalizeVersion(s.deps.CurrentVersion()),
			Error:          err.Error(),
		}
	}

	current := normalizeVersion(s.deps.CurrentVersion())
	updateAvailable, err := isNewerVersion(manifest.Version, current)
	if err != nil {
		return domainupdate.CheckResult{
			CurrentVersion: current,
			LatestVersion:  normalizeVersion(manifest.Version),
			DownloadURL:    strings.TrimSpace(manifest.LinuxURL),
			Notes:          manifest.Notes,
			Error:          err.Error(),
		}
	}

	return domainupdate.CheckResult{
		CurrentVersion:  current,
		LatestVersion:   normalizeVersion(manifest.Version),
		UpdateAvailable: updateAvailable,
		DownloadURL:     strings.TrimSpace(manifest.LinuxURL),
		Notes:           manifest.Notes,
	}
}

func (s *Service) DownloadLatest(ctx context.Context, manifestURL string) domainupdate.DownloadResult {
	check := s.CheckForUpdate(ctx, manifestURL)
	if check.Error != "" {
		return domainupdate.DownloadResult{
			CurrentVersion: check.CurrentVersion,
			LatestVersion:  check.LatestVersion,
			Notes:          check.Notes,
			Error:          check.Error,
		}
	}
	if !check.UpdateAvailable {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: false,
			Notes:           check.Notes,
		}
	}
	if s.deps.GetLicenseKey == nil {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			Notes:           check.Notes,
			Error:           "license service is unavailable",
		}
	}

	licenseResult := s.deps.GetLicenseKey(ctx)
	if licenseResult.Error != "" {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			Notes:           check.Notes,
			Error:           licenseResult.Error,
		}
	}
	if strings.TrimSpace(licenseResult.LicenseKey) == "" {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			Notes:           check.Notes,
			Error:           "license key is required to download updates",
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, check.DownloadURL, nil)
	if err != nil {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			Notes:           check.Notes,
			Error:           fmt.Sprintf("create update download request: %v", err),
		}
	}
	req.Header.Set("X-Phant-License", licenseResult.LicenseKey)

	resp, err := s.deps.HTTPClient().Do(req)
	if err != nil {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			Notes:           check.Notes,
			Error:           fmt.Sprintf("download update: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			StatusCode:      resp.StatusCode,
			FinalURL:        resp.Request.URL.String(),
			Notes:           check.Notes,
			Error:           fmt.Sprintf("update endpoint returned status %d", resp.StatusCode),
		}
	}

	file, err := os.CreateTemp("", "phant-update-*.AppImage")
	if err != nil {
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			StatusCode:      resp.StatusCode,
			FinalURL:        resp.Request.URL.String(),
			Notes:           check.Notes,
			Error:           fmt.Sprintf("create temp update file: %v", err),
		}
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	if err != nil {
		_ = os.Remove(file.Name())
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			StatusCode:      resp.StatusCode,
			FinalURL:        resp.Request.URL.String(),
			Notes:           check.Notes,
			Error:           fmt.Sprintf("write update payload: %v", err),
		}
	}

	if err := os.Chmod(file.Name(), 0o755); err != nil {
		_ = os.Remove(file.Name())
		return domainupdate.DownloadResult{
			CurrentVersion:  check.CurrentVersion,
			LatestVersion:   check.LatestVersion,
			UpdateAvailable: true,
			StatusCode:      resp.StatusCode,
			FinalURL:        resp.Request.URL.String(),
			BytesWritten:    written,
			Notes:           check.Notes,
			Error:           fmt.Sprintf("set update file permissions: %v", err),
		}
	}

	return domainupdate.DownloadResult{
		CurrentVersion:  check.CurrentVersion,
		LatestVersion:   check.LatestVersion,
		UpdateAvailable: true,
		Downloaded:      true,
		FilePath:        file.Name(),
		FinalURL:        resp.Request.URL.String(),
		StatusCode:      resp.StatusCode,
		BytesWritten:    written,
		Notes:           check.Notes,
	}
}

func (s *Service) fetchManifest(ctx context.Context, manifestURL string) (domainupdate.Manifest, error) {
	resolvedManifestURL := strings.TrimSpace(manifestURL)
	if resolvedManifestURL == "" {
		resolvedManifestURL = DefaultManifestURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolvedManifestURL, nil)
	if err != nil {
		return domainupdate.Manifest{}, fmt.Errorf("create update manifest request: %w", err)
	}

	resp, err := s.deps.HTTPClient().Do(req)
	if err != nil {
		return domainupdate.Manifest{}, fmt.Errorf("fetch update manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return domainupdate.Manifest{}, fmt.Errorf("update manifest returned status %d", resp.StatusCode)
	}

	var manifest domainupdate.Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return domainupdate.Manifest{}, fmt.Errorf("decode update manifest: %w", err)
	}

	manifest.Version = normalizeVersion(manifest.Version)
	manifest.LinuxURL = strings.TrimSpace(manifest.LinuxURL)
	if manifest.Version == "" {
		return domainupdate.Manifest{}, fmt.Errorf("update manifest version is empty")
	}
	if manifest.LinuxURL == "" {
		return domainupdate.Manifest{}, fmt.Errorf("update manifest linux_url is empty")
	}

	return manifest, nil
}

func isNewerVersion(latest string, current string) (bool, error) {
	latestParts, err := parseVersion(latest)
	if err != nil {
		return false, fmt.Errorf("invalid latest version %q: %w", latest, err)
	}
	currentParts, err := parseVersion(current)
	if err != nil {
		return false, fmt.Errorf("invalid current version %q: %w", current, err)
	}

	for i := 0; i < len(latestParts) || i < len(currentParts); i++ {
		latestPart := 0
		currentPart := 0
		if i < len(latestParts) {
			latestPart = latestParts[i]
		}
		if i < len(currentParts) {
			currentPart = currentParts[i]
		}
		if latestPart > currentPart {
			return true, nil
		}
		if latestPart < currentPart {
			return false, nil
		}
	}

	return false, nil
}

func parseVersion(v string) ([]int, error) {
	normalized := normalizeVersion(v)
	if normalized == "" {
		return nil, fmt.Errorf("version is empty")
	}

	parts := strings.Split(normalized, ".")
	parsed := make([]int, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("version contains empty segment")
		}
		number, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("version segment %q is not numeric", part)
		}
		parsed = append(parsed, number)
	}

	return parsed, nil
}

func normalizeVersion(v string) string {
	normalized := strings.TrimSpace(v)
	normalized = strings.TrimPrefix(normalized, "v")
	if idx := strings.IndexAny(normalized, "-+"); idx >= 0 {
		normalized = normalized[:idx]
	}
	return normalized
}

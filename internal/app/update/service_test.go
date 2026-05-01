package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	domainlicense "phant/internal/domain/license"
	domainupdate "phant/internal/domain/update"
)

func TestServiceCheckForUpdate(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/update.json" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{"version":"1.1.0","linux_url":"%s/download","notes":"Bug fixes"}`, server.URL)
	}))
	defer server.Close()

	svc := NewService(Dependencies{
		CurrentVersion: func() string { return "1.0.0" },
		HTTPClient:     func() *http.Client { return server.Client() },
	})

	got := svc.CheckForUpdate(context.Background(), server.URL+"/update.json")
	if got.Error != "" {
		t.Fatalf("CheckForUpdate(...) error = %q", got.Error)
	}
	if !got.UpdateAvailable {
		t.Fatalf("CheckForUpdate(...) updateAvailable = false, want true")
	}
}

func TestServiceDownloadLatestFollowsRedirectWithLicense(t *testing.T) {
	const expectedLicense = "PHANT-KEY-1234"
	const expectedPlatform = "linux"
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/update.json":
			fmt.Fprintf(w, `{"version":"1.2.0","linux_url":"%s/download","notes":"Release note"}`, server.URL)
		case "/download":
			if r.Header.Get("X-Phant-License") != expectedLicense {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if r.Header.Get("X-Phant-Platform") != expectedPlatform {
				http.Error(w, "missing platform header", http.StatusBadRequest)
				return
			}
			http.Redirect(w, r, "/artifact", http.StatusFound)
		case "/artifact":
			_, _ = w.Write([]byte("appimage-bytes"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewService(Dependencies{
		CurrentVersion: func() string { return "1.0.0" },
		Platform:       func() string { return "linux" },
		GetLicenseKey: func(context.Context) domainlicense.KeyResult {
			return domainlicense.KeyResult{LicenseKey: expectedLicense}
		},
		HTTPClient:         func() *http.Client { return server.Client() },
		DownloadHTTPClient: func() *http.Client { return server.Client() },
	})

	result := svc.DownloadLatest(context.Background(), server.URL+"/update.json")
	if result.Error != "" {
		t.Fatalf("DownloadLatest(...) error = %q", result.Error)
	}
	if !result.Downloaded {
		t.Fatalf("DownloadLatest(...) downloaded = false, want true")
	}
	if result.FinalURL != server.URL+"/artifact" {
		t.Fatalf("DownloadLatest(...) finalURL = %q, want %q", result.FinalURL, server.URL+"/artifact")
	}

	data, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(result.FilePath) })
	if string(data) != "appimage-bytes" {
		t.Fatalf("downloaded content = %q, want %q", string(data), "appimage-bytes")
	}
}

func TestServiceInstallDownloadedLinux(t *testing.T) {
	var receivedPath string
	svc := NewService(Dependencies{
		InstallDownloaded: func(_ context.Context, path string) domainupdate.InstallResult {
			receivedPath = path
			return domainupdate.InstallResult{
				Installed:  true,
				TargetPath: "/opt/phant/phant.AppImage",
				Message:    "ok",
			}
		},
	})

	result := svc.InstallDownloaded(context.Background(), "/tmp/update.AppImage")
	if result.Error != "" {
		t.Fatalf("InstallDownloaded(...) error = %q", result.Error)
	}
	if !result.Installed {
		t.Fatalf("InstallDownloaded(...) installed = false, want true")
	}
	if receivedPath != "/tmp/update.AppImage" {
		t.Fatalf("InstallDownloaded(...) forwarded path = %q, want %q", receivedPath, "/tmp/update.AppImage")
	}
}

func TestServiceInstallDownloadedRejectsNonLinux(t *testing.T) {
	svc := NewService(Dependencies{})

	result := svc.InstallDownloaded(context.Background(), "/tmp/update.AppImage")
	if result.Error == "" {
		t.Fatalf("InstallDownloaded(...) expected unavailable installer error")
	}
}

func TestNewServiceSetsDefaultHTTPTimeouts(t *testing.T) {
	svc := NewService(Dependencies{})

	if svc.deps.HTTPClient().Timeout != defaultManifestHTTPTimeout {
		t.Fatalf("HTTPClient timeout = %v, want %v", svc.deps.HTTPClient().Timeout, defaultManifestHTTPTimeout)
	}
	if svc.deps.DownloadHTTPClient().Timeout != defaultDownloadHTTPTimeout {
		t.Fatalf("DownloadHTTPClient timeout = %v, want %v", svc.deps.DownloadHTTPClient().Timeout, defaultDownloadHTTPTimeout)
	}
}

func TestServiceCheckForUpdateRejectsOversizedManifest(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/update.json" {
			http.NotFound(w, r)
			return
		}
		payload := `{"version":"1.1.0","linux_url":"` + server.URL + `/download","notes":"` + strings.Repeat("x", maxManifestBytes) + `"}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer server.Close()

	svc := NewService(Dependencies{
		CurrentVersion: func() string { return "1.0.0" },
		HTTPClient:     func() *http.Client { return server.Client() },
	})

	got := svc.CheckForUpdate(context.Background(), server.URL+"/update.json")
	if got.Error == "" {
		t.Fatalf("CheckForUpdate(...) expected oversized manifest error")
	}
}

func TestServiceDownloadLatestRejectsOversizedPayload(t *testing.T) {
	const expectedLicense = "PHANT-KEY-1234"
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/update.json":
			fmt.Fprintf(w, `{"version":"1.2.0","linux_url":"%s/download","notes":"Release note"}`, server.URL)
		case "/download":
			http.Redirect(w, r, "/artifact", http.StatusFound)
		case "/artifact":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", maxDownloadBytes+1))
			_, _ = w.Write([]byte("oversized"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	svc := NewService(Dependencies{
		CurrentVersion: func() string { return "1.0.0" },
		Platform:       func() string { return "linux" },
		GetLicenseKey: func(context.Context) domainlicense.KeyResult {
			return domainlicense.KeyResult{LicenseKey: expectedLicense}
		},
		HTTPClient:         func() *http.Client { return server.Client() },
		DownloadHTTPClient: func() *http.Client { return server.Client() },
	})

	result := svc.DownloadLatest(context.Background(), server.URL+"/update.json")
	if result.Error == "" {
		t.Fatalf("DownloadLatest(...) expected oversized payload error")
	}
}

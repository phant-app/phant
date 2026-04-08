package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUpdateServiceCheckForUpdate(t *testing.T) {
	originalVersion := BuildVersion
	BuildVersion = "1.0.0"
	t.Cleanup(func() {
		BuildVersion = originalVersion
	})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/update.json" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{"version":"1.1.0","linux_url":"%s/download","notes":"Bug fixes"}`, server.URL)
	}))
	defer server.Close()

	service := NewUpdateService()
	result := service.CheckForUpdate(server.URL + "/update.json")

	if result.Error != "" {
		t.Fatalf("CheckForUpdate(...) error = %q", result.Error)
	}
	if !result.UpdateAvailable {
		t.Fatalf("CheckForUpdate(...) updateAvailable = false, want true")
	}
	if result.LatestVersion != "1.1.0" {
		t.Fatalf("CheckForUpdate(...) latestVersion = %q, want %q", result.LatestVersion, "1.1.0")
	}
}

func TestUpdateServiceDownloadLatestFollowsRedirectWithLicense(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	originalVersion := BuildVersion
	BuildVersion = "1.0.0"
	t.Cleanup(func() {
		BuildVersion = originalVersion
	})

	const expectedLicense = "PHANT-KEY-1234"
	saveResult := NewLicenseService().SaveLicenseKey(expectedLicense)
	if !saveResult.Success {
		t.Fatalf("SaveLicenseKey(...) success = false, error=%q", saveResult.Error)
	}

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
			http.Redirect(w, r, "/artifact", http.StatusFound)
		case "/artifact":
			_, _ = w.Write([]byte("appimage-bytes"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := NewUpdateService()
	result := service.DownloadLatest(server.URL + "/update.json")

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
	t.Cleanup(func() {
		_ = os.Remove(result.FilePath)
	})

	if string(data) != "appimage-bytes" {
		t.Fatalf("downloaded file content = %q, want %q", string(data), "appimage-bytes")
	}
}

func TestUpdateServiceDownloadLatestRequiresLicense(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	originalVersion := BuildVersion
	BuildVersion = "1.0.0"
	t.Cleanup(func() {
		BuildVersion = originalVersion
	})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/update.json" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{"version":"1.1.0","linux_url":"%s/download","notes":"Release note"}`, server.URL)
	}))
	defer server.Close()

	service := NewUpdateService()
	result := service.DownloadLatest(server.URL + "/update.json")

	if result.Error == "" {
		t.Fatalf("DownloadLatest(...) expected license key error")
	}
}

package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	domainlicense "phant/internal/domain/license"
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

	svc := NewService(Dependencies{
		CurrentVersion: func() string { return "1.0.0" },
		GetLicenseKey: func(context.Context) domainlicense.KeyResult {
			return domainlicense.KeyResult{LicenseKey: expectedLicense}
		},
		HTTPClient: func() *http.Client { return server.Client() },
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

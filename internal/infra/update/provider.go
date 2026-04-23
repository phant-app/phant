package update

import (
	"context"
	"net/http"

	domainupdate "phant/internal/domain/update"
)

// Provider defines OS-specific update plumbing.
type Provider interface {
	Platform() string
	HTTPClient() *http.Client
	InstallDownloaded(ctx context.Context, downloadedPath string) domainupdate.InstallResult
}

package update

import (
	"context"
	"fmt"
	"net/http"
	"time"

	domainupdate "phant/internal/domain/update"
)

type unsupportedProvider struct {
	platform string
}

func newUnsupportedProvider(platform string) Provider {
	return unsupportedProvider{platform: platform}
}

func (p unsupportedProvider) Platform() string {
	return p.platform
}

func (p unsupportedProvider) HTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func (p unsupportedProvider) InstallDownloaded(context.Context, string) domainupdate.InstallResult {
	return domainupdate.InstallResult{
		Error: fmt.Sprintf("install update is currently unsupported on %s", p.platform),
	}
}

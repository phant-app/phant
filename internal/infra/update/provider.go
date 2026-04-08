package update

import (
	"net/http"
	"time"
)

type Provider interface {
	Platform() string
	HTTPClient() *http.Client
}

type defaultProvider struct {
	platform string
	client   *http.Client
}

func NewProviderForCurrentOS() Provider {
	return defaultProvider{
		platform: runtimePlatform(),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p defaultProvider) Platform() string {
	return p.platform
}

func (p defaultProvider) HTTPClient() *http.Client {
	return p.client
}

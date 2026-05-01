package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	domainsettings "phant/internal/domain/settings"
)

type Provider interface {
	Platform() string
	Load(context.Context) (domainsettings.AppSettings, error)
	Save(context.Context, domainsettings.AppSettings) error
}

type fileProvider struct {
	platform string
}

func NewFileProvider() Provider {
	return fileProvider{platform: runtimePlatform()}
}

func (p fileProvider) Platform() string {
	return p.platform
}

func (p fileProvider) Load(context.Context) (domainsettings.AppSettings, error) {
	path, err := settingsFilePath()
	if err != nil {
		return domainsettings.AppSettings{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return domainsettings.AppSettings{}, nil
	}
	if err != nil {
		return domainsettings.AppSettings{}, fmt.Errorf("read settings file: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return domainsettings.AppSettings{}, nil
	}

	var settings domainsettings.AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return domainsettings.AppSettings{}, fmt.Errorf("decode settings file: %w", err)
	}

	return settings, nil
}

func (p fileProvider) Save(_ context.Context, settings domainsettings.AppSettings) error {
	path, err := settingsFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings file: %w", err)
	}
	data = append(data, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write temp settings file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace settings file: %w", err)
	}

	return nil
}

func settingsFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(configDir, "phant", "settings.json"), nil
}

// Package config loads cf-genimage settings from the environment.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the Cloudflare credentials.
type Config struct {
	AccountID string
	APIToken  string
}

// Load reads CF_ACCOUNT_ID and CF_API_TOKEN from the environment. A .env file
// in the working directory is loaded first (if present) as a development
// convenience; real environment variables take precedence.
func Load() (Config, error) {
	loadDotEnv(".env")
	cfg := Config{
		AccountID: strings.TrimSpace(os.Getenv("CF_ACCOUNT_ID")),
		APIToken:  strings.TrimSpace(os.Getenv("CF_API_TOKEN")),
	}
	if cfg.AccountID == "" || cfg.APIToken == "" {
		return Config{}, errors.New("missing Cloudflare credentials: set CF_ACCOUNT_ID and CF_API_TOKEN environment variables")
	}
	return cfg, nil
}

// loadDotEnv parses KEY=VALUE lines without external dependencies.
// A missing file is not an error; existing environment wins over file values.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if k == "" || os.Getenv(k) != "" {
			continue
		}
		_ = os.Setenv(k, strings.Trim(strings.TrimSpace(v), `"'`))
	}
}

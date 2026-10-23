package config_test

import (
	"testing"

	"cf-genimage/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("CF_ACCOUNT_ID", "acc123")
	t.Setenv("CF_API_TOKEN", "tok456")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AccountID != "acc123" || cfg.APIToken != "tok456" {
		t.Fatalf("config = %+v, want acc123/tok456", cfg)
	}
}

func TestLoad_Missing(t *testing.T) {
	t.Setenv("CF_ACCOUNT_ID", "")
	t.Setenv("CF_API_TOKEN", "")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

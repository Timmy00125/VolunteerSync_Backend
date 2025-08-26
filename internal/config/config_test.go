package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port == 0 || cfg.Host == "" {
		t.Fatalf("invalid defaults: host=%s port=%d", cfg.Host, cfg.Port)
	}
}

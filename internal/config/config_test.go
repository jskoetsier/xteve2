// internal/config/config_test.go
package config_test

import (
	"testing"

	"xteve/internal/config"
	"xteve/internal/storage"
)

func TestDefaults(t *testing.T) {
	cfg := config.Default()

	if cfg.Port != 34400 {
		t.Errorf("default port = %d, want 34400", cfg.Port)
	}
	if cfg.TunerCount != 1 {
		t.Errorf("default tuner count = %d, want 1", cfg.TunerCount)
	}
	if cfg.AuthEnabled {
		t.Error("auth should be disabled by default")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	original := config.Default()
	original.Port = 9000
	original.AuthEnabled = true

	if err := config.Save(s, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load(s)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Port != 9000 {
		t.Errorf("port = %d, want 9000", loaded.Port)
	}
	if !loaded.AuthEnabled {
		t.Error("AuthEnabled should be true")
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	cfg, err := config.Load(s)
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}

	if cfg.Port != 34400 {
		t.Errorf("port = %d, want default 34400", cfg.Port)
	}
}

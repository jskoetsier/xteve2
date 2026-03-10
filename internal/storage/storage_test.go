// internal/storage/storage_test.go
package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"xteve/internal/storage"
)

type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	original := TestData{Name: "hello", Value: 42}
	if err := s.Save("test.json", original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var loaded TestData
	if err := s.Load("test.json", &loaded); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded != original {
		t.Errorf("got %+v, want %+v", loaded, original)
	}
}

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	var v TestData
	err := s.Load("missing.json", &v)
	if !storage.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	dirs := []string{"cache", "backup", "temp", "img-cache", "img-upload"}
	if err := s.EnsureDirs(dirs...); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(dir, d)); err != nil {
			t.Errorf("dir %q not created: %v", d, err)
		}
	}
}

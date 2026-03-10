// internal/storage/storage.go
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var errNotFound = errors.New("not found")

// IsNotFound reports whether err is a not-found error from Load.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// Storage manages JSON file persistence in a directory.
type Storage struct {
	dir string
	mu  sync.Mutex
}

// New returns a Storage rooted at dir.
func New(dir string) *Storage {
	return &Storage{dir: dir}
}

// Dir returns the root directory.
func (s *Storage) Dir() string {
	return s.dir
}

// Save writes v as JSON to filename inside the storage directory.
func (s *Storage) Save(filename string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, filename)
	return os.WriteFile(path, data, 0o644)
}

// Load reads filename from the storage directory and unmarshals it into v.
// Returns an IsNotFound error if the file does not exist.
func (s *Storage) Load(filename string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, filename)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return errNotFound
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// EnsureDirs creates subdirectories inside the storage directory.
func (s *Storage) EnsureDirs(names ...string) error {
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(s.dir, name), 0o755); err != nil {
			return err
		}
	}
	return nil
}

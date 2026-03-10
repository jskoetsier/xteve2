// internal/api/api_test.go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"xteve/internal/api"
	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/storage"
	"xteve/internal/xepg"
)

func newTestAPI(t *testing.T) *api.API {
	t.Helper()
	s := storage.New(t.TempDir())
	cfg := config.Default()
	db := xepg.NewDB()
	buf := buffer.New(buffer.Config{TunerCount: cfg.TunerCount})
	return api.New(api.Config{
		Storage:  s,
		Settings: cfg,
		XEPG:     db,
		Buffer:   buf,
	})
}

func TestStatusEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["version"] == nil {
		t.Error("response missing 'version' field")
	}
}

func TestSettingsGetEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var s config.Settings
	if err := json.NewDecoder(w.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Port != 34400 {
		t.Errorf("port = %d, want 34400", s.Port)
	}
}

func TestChannelsEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/channels", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

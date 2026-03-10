// internal/auth/auth_test.go
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"xteve/internal/auth"
)

func TestDisabledAllowsAll(t *testing.T) {
	a := auth.New(auth.Config{Enabled: false})

	handler := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("disabled auth: status = %d, want 200", w.Code)
	}
}

func TestSetAndVerifyPassword(t *testing.T) {
	a := auth.New(auth.Config{Enabled: true})

	if err := a.SetPassword("secret123"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if !a.CheckPassword("secret123") {
		t.Error("CheckPassword: correct password rejected")
	}
	if a.CheckPassword("wrong") {
		t.Error("CheckPassword: wrong password accepted")
	}
}

func TestUnauthenticatedReturns401(t *testing.T) {
	a := auth.New(auth.Config{Enabled: true})
	_ = a.SetPassword("secret")

	handler := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated: status = %d, want 401", w.Code)
	}
}

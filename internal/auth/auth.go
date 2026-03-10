// internal/auth/auth.go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config holds auth configuration.
type Config struct {
	Enabled      bool
	PasswordHash string // bcrypt hash; empty means not set
}

// Auth manages optional session-based authentication.
type Auth struct {
	mu       sync.RWMutex
	cfg      Config
	sessions map[string]time.Time // token → expiry
}

const (
	cookieName = "xteve_session"
	sessionTTL = 24 * time.Hour
)

// New creates an Auth from config.
func New(cfg Config) *Auth {
	return &Auth{
		cfg:      cfg,
		sessions: make(map[string]time.Time),
	}
}

// SetPassword hashes and stores a new password.
func (a *Auth) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.PasswordHash = string(hash)
	a.mu.Unlock()
	return nil
}

// CheckPassword reports whether password matches the stored hash.
func (a *Auth) CheckPassword(password string) bool {
	a.mu.RLock()
	hash := a.cfg.PasswordHash
	a.mu.RUnlock()
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Login creates a session token for the given request and sets a cookie.
func (a *Auth) Login(w http.ResponseWriter) string {
	token := randomToken()
	a.mu.Lock()
	a.sessions[token] = time.Now().Add(sessionTTL)
	a.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// Logout invalidates the session from the request cookie.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, MaxAge: -1, Path: "/"})
}

// Middleware returns an HTTP middleware that enforces auth when enabled.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.RLock()
		enabled := a.cfg.Enabled
		a.mu.RUnlock()

		if !enabled {
			next.ServeHTTP(w, r)
			return
		}

		if c, err := r.Cookie(cookieName); err == nil {
			a.mu.RLock()
			expiry, ok := a.sessions[c.Value]
			a.mu.RUnlock()

			if ok && time.Now().Before(expiry) {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func randomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

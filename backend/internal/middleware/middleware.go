package middleware

import (
	"context"
	"michealenterprise/internal/models"
	"michealenterprise/internal/repository"
	"net/http"
	"sync"
	"time"
)

type key struct{}

func User(r *http.Request) models.User { u, _ := r.Context().Value(key{}).(models.User); return u }
func Auth(repo *repository.Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, e := r.Cookie("me_session"); e == nil {
			if u, e := repo.Session(r.Context(), c.Value); e == nil {
				r = r.WithContext(context.WithValue(r.Context(), key{}, u))
			}
		}
		next.ServeHTTP(w, r)
	})
}
func Security(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" && r.Header.Get("Origin") != origin {
			http.Error(w, "origin rejected", 403)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
		next.ServeHTTP(w, r)
	})
}

type entry struct {
	n     int
	reset time.Time
}
type Limiter struct {
	mu      sync.Mutex
	entries map[string]entry
}

func NewLimiter() *Limiter { return &Limiter{entries: map[string]entry{}} }
func (l *Limiter) Allow(key string, limit int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for k, e := range l.entries {
		if now.After(e.reset) {
			delete(l.entries, k)
		}
	}
	e := l.entries[key]
	if e.reset.IsZero() {
		e.reset = now.Add(time.Minute)
	}
	e.n++
	l.entries[key] = e
	return e.n <= limit
}

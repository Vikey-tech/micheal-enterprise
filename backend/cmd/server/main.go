package main

import (
	"context"
	"log/slog"
	"michealenterprise/internal/handlers"
	"michealenterprise/internal/repository"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func main() {
	origin := env("APP_ORIGIN", "http://localhost:8080")
	u, e := url.Parse(origin)
	if e != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.Path != "" {
		slog.Error("APP_ORIGIN must be a full origin without a trailing slash")
		os.Exit(1)
	}
	secure := env("COOKIE_SECURE", "true") == "true"
	if secure && u.Scheme != "https" {
		slog.Error("Secure cookies require HTTPS; use COOKIE_SECURE=false only for local development")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	repo, e := repository.Open(ctx, os.Getenv("DATABASE_URL"))
	if e != nil {
		slog.Error("database unavailable", "error", e)
		os.Exit(1)
	}
	defer repo.DB.Close()
	if e = repo.Migrate(ctx); e != nil {
		slog.Error("migration failed", "error", e)
		os.Exit(1)
	}
	uploads := env("UPLOAD_DIR", "/data/uploads")
	if e = os.MkdirAll(uploads, 0750); e != nil {
		slog.Error("upload directory unavailable", "error", e)
		os.Exit(1)
	}
	api := handlers.New(repo, origin, uploads, env("IMAGE_OPTIMIZER", "/app/tools/optimize_images.py"), secure)
	server := &http.Server{Addr: env("LISTEN_ADDR", ":8081"), Handler: api.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	go func() {
		<-ctx.Done()
		stop, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(stop)
	}()
	// Sessions are fixed-duration; remove expired rows without touching active sessions.
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				repo.DB.Exec(ctx, "DELETE FROM sessions WHERE expires_at<now()")
			}
		}
	}()
	slog.Info("API listening", "address", server.Addr)
	if e = server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		slog.Error("server failed", "error", e)
		os.Exit(1)
	}
}

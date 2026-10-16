// setup is an operator-only CLI, never a public HTTP endpoint.
package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"michealenterprise/internal/repository"
	"michealenterprise/internal/service"
	"os"
	"strings"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, e := repository.Open(ctx, os.Getenv("DATABASE_URL"))
	if e != nil {
		return e
	}
	defer r.DB.Close()
	if e = r.Migrate(ctx); e != nil {
		return e
	}
	tx, e := r.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(827451)"); e != nil {
		return e
	}
	var count int
	if e = tx.QueryRow(ctx, "SELECT count(*) FROM users WHERE role IN ('owner','manager')").Scan(&count); e != nil {
		return e
	}
	if count != 0 {
		return fmt.Errorf("administrators already exist; setup refused")
	}
	for _, role := range []string{"owner", "manager"} {
		prefix := strings.ToUpper(role)
		email := strings.ToLower(strings.TrimSpace(os.Getenv(prefix + "_EMAIL")))
		password := os.Getenv(prefix + "_PASSWORD")
		name := os.Getenv(prefix + "_NAME")
		if !service.Email(email) || !service.Password(password) || len(name) < 2 || len(name) > 100 {
			return fmt.Errorf("set valid %s_EMAIL, %s_NAME and %s_PASSWORD (12–72 bytes)", prefix, prefix, prefix)
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(password), 12)
		if e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "INSERT INTO users(name,email,password_hash,role) VALUES($1,$2,$3,$4)", name, email, string(hash), role); e != nil {
			return e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return e
	}
	fmt.Println("Owner and manager created. Remove setup password variables now.")
	return nil
}

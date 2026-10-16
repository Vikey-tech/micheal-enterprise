package handlers

import (
	"context"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"michealenterprise/internal/middleware"
	"michealenterprise/internal/models"
	"michealenterprise/internal/repository"
	"michealenterprise/internal/service"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type API struct {
	Repo                       *repository.Repository
	Origin, Uploads, Optimizer string
	Secure                     bool
	limit                      *middleware.Limiter
	sockets                    sync.Map
	dummy                      []byte
}

func New(repo *repository.Repository, origin, uploads, optimizer string, secure bool) *API {
	dummy, _ := bcrypt.GenerateFromPassword([]byte(service.Token()), 12)
	return &API{Repo: repo, Origin: origin, Uploads: uploads, Optimizer: optimizer, Secure: secure, limit: middleware.NewLimiter(), dummy: dummy}
}
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, 400, "Invalid request fields")
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		fail(w, 400, "Send one JSON object")
		return false
	}
	return true
}
func (a *API) dbError(w http.ResponseWriter, err error) {
	slog.Error("database request failed", "error", err)
	fail(w, 500, "Unable to complete request. Please try again.")
}
func (a *API) required(area string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := middleware.User(r)
		if u.ID == "" {
			fail(w, 401, "Please sign in")
			return
		}
		if area != "" && !service.Can(u, area) {
			fail(w, 403, "Permission denied")
			return
		}
		h(w, r)
	}
}
func (a *API) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.Repo.DB.Ping(ctx); err != nil {
			fail(w, 503, "Database unavailable")
			return
		}
		JSON(w, 200, map[string]string{"status": "ok"})
	})
	m.HandleFunc("POST /api/auth/register", a.register)
	m.HandleFunc("POST /api/auth/login", a.login)
	m.HandleFunc("POST /api/auth/logout", a.logout)
	m.HandleFunc("GET /api/auth/me", a.required("", func(w http.ResponseWriter, r *http.Request) { JSON(w, 200, middleware.User(r)) }))
	m.HandleFunc("GET /api/vehicles", a.vehicles)
	m.HandleFunc("GET /api/vehicles/{id}", a.vehicle)
	m.HandleFunc("GET /api/admin/vehicles", a.required("vehicles", a.vehicles))
	m.HandleFunc("POST /api/admin/vehicles", a.required("vehicles", a.saveVehicle))
	m.HandleFunc("PUT /api/admin/vehicles/{id}", a.required("vehicles", a.saveVehicle))
	m.HandleFunc("POST /api/admin/uploads", a.required("", a.upload))
	m.HandleFunc("GET /api/profile", func(w http.ResponseWriter, r *http.Request) {
		p, e := a.Repo.Profile(r.Context())
		if e != nil {
			a.dbError(w, e)
			return
		}
		JSON(w, 200, p)
	})
	m.HandleFunc("PUT /api/admin/profile", a.required("owner-only", a.profile))
	m.HandleFunc("GET /api/admin/stats", a.required("", a.stats))
	m.HandleFunc("GET /api/admin/manager", a.required("owner-only", a.manager))
	m.HandleFunc("PUT /api/admin/manager", a.required("owner-only", a.manager))
	m.HandleFunc("POST /api/enquiries", a.enquiry)
	m.HandleFunc("GET /api/admin/enquiries", a.required("enquiries", a.enquiries))
	m.HandleFunc("PUT /api/admin/enquiries/{id}", a.required("enquiries", a.enquiries))
	m.HandleFunc("GET /api/conversations", a.required("", a.conversations))
	m.HandleFunc("POST /api/conversations", a.required("", a.conversations))
	m.HandleFunc("PUT /api/conversations/{id}", a.required("chat", a.conversationStatus))
	m.HandleFunc("GET /api/conversations/{id}/messages", a.required("", a.messages))
	m.HandleFunc("POST /api/conversations/{id}/messages", a.required("", a.messages))
	m.HandleFunc("GET /api/ws", a.required("", a.ws))
	// Only generated WebP filenames may be served; directory listing is disabled.
	m.HandleFunc("GET /uploads/{name}", func(w http.ResponseWriter, r *http.Request) {
		if !service.Image("/uploads/" + r.PathValue("name")) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/webp")
		http.ServeFile(w, r, filepath.Join(a.Uploads, r.PathValue("name")))
	})
	return middleware.Security(a.Origin, middleware.Auth(a.Repo, m))
}
func (a *API) limited(w http.ResponseWriter, r *http.Request, kind string, n int) bool {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !a.limit.Allow(kind+ip, n) {
		fail(w, 429, "Too many requests. Please wait a minute.")
		return true
	}
	return false
}
func (a *API) register(w http.ResponseWriter, r *http.Request) {
	if a.limited(w, r, "auth", 10) {
		return
	}
	var b struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
	if len(b.Name) < 2 || len(b.Name) > 100 || !service.Email(b.Email) || !service.Password(b.Password) {
		fail(w, 400, "Use a name, valid email and a password of 12–72 bytes")
		return
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(b.Password), 12)
	if e != nil {
		a.dbError(w, e)
		return
	}
	var id string
	e = a.Repo.DB.QueryRow(r.Context(), "INSERT INTO users(name,email,password_hash,role) VALUES($1,$2,$3,'customer') ON CONFLICT(email) DO NOTHING RETURNING id", b.Name, b.Email, string(hash)).Scan(&id)
	if e != nil {
		fail(w, 409, "Unable to register with this email")
		return
	}
	a.session(w, r, id)
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	if a.limited(w, r, "auth", 10) {
		return
	}
	var b struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &b) {
		return
	}
	u, hash, e := a.Repo.UserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(b.Email)))
	if e != nil {
		bcrypt.CompareHashAndPassword(a.dummy, []byte(b.Password))
		fail(w, 401, "Email or password is incorrect")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(b.Password)) != nil {
		fail(w, 401, "Email or password is incorrect")
		return
	}
	a.session(w, r, u.ID)
}
func (a *API) session(w http.ResponseWriter, r *http.Request, id string) {
	token := service.Token()
	if e := a.Repo.NewSession(r.Context(), id, token); e != nil {
		a.dbError(w, e)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "me_session", Value: token, Path: "/", HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	u, e := a.Repo.Session(r.Context(), token)
	if e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, u)
}
func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if c, e := r.Cookie("me_session"); e == nil {
		if _, e = a.Repo.DB.Exec(r.Context(), "DELETE FROM sessions WHERE token_hash=$1", repository.HashToken(c.Value)); e != nil {
			a.dbError(w, e)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: "me_session", Value: "", Path: "/", HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	JSON(w, 200, map[string]bool{"ok": true})
}
func (a *API) vehicles(w http.ResponseWriter, r *http.Request) {
	vs, e := a.Repo.Vehicles(r.Context(), strings.HasPrefix(r.URL.Path, "/api/admin/"))
	if e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, vs)
}
func (a *API) vehicle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !service.UUID(id) {
		fail(w, 404, "Vehicle not found")
		return
	}
	vs, e := a.Repo.Vehicles(r.Context(), false)
	if e != nil {
		a.dbError(w, e)
		return
	}
	for _, v := range vs {
		if v.ID == id {
			JSON(w, 200, v)
			return
		}
	}
	fail(w, 404, "Vehicle not found")
}
func (a *API) saveVehicle(w http.ResponseWriter, r *http.Request) {
	var v models.Vehicle
	if !decode(w, r, &v) {
		return
	}
	if e := service.Vehicle(v); e != nil {
		fail(w, 400, e.Error())
		return
	}
	if v.Images == nil {
		v.Images = []string{}
	}
	b, _ := json.Marshal(v)
	id := r.PathValue("id")
	if r.Method == "POST" {
		e := a.Repo.DB.QueryRow(r.Context(), "INSERT INTO vehicles(data) VALUES($1) RETURNING id", b).Scan(&id)
		if e != nil {
			a.dbError(w, e)
			return
		}
	} else {
		if !service.UUID(id) {
			fail(w, 400, "Invalid vehicle ID")
			return
		}
		tag, e := a.Repo.DB.Exec(r.Context(), "UPDATE vehicles SET data=$1 WHERE id=$2", b, id)
		if e != nil {
			a.dbError(w, e)
			return
		}
		if tag.RowsAffected() == 0 {
			fail(w, 404, "Vehicle not found")
			return
		}
	}
	v.ID = id
	JSON(w, 200, v)
}
func (a *API) profile(w http.ResponseWriter, r *http.Request) {
	var p models.Profile
	if !decode(w, r, &p) {
		return
	}
	if len(p.Name) < 2 || len(p.Name) > 100 || len(p.Bio) > 5000 || len(p.Phone) > 30 || len(p.WhatsApp) > 20 || len(p.Address) > 500 || !service.Image(p.Image) || (p.Email != "" && !service.Email(p.Email)) {
		fail(w, 400, "Check profile fields")
		return
	}
	for _, c := range p.WhatsApp {
		if c < '0' || c > '9' {
			fail(w, 400, "WhatsApp must contain country code and digits only")
			return
		}
	}
	b, _ := json.Marshal(p)
	if _, e := a.Repo.DB.Exec(r.Context(), "UPDATE profile SET data=$1 WHERE id=1", b); e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, p)
}
func (a *API) manager(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		var p map[string]bool
		if !decode(w, r, &p) {
			return
		}
		for k := range p {
			if k != "vehicles" && k != "chat" && k != "enquiries" {
				fail(w, 400, "Unknown permission")
				return
			}
		}
		b, _ := json.Marshal(p)
		if _, e := a.Repo.DB.Exec(r.Context(), "UPDATE users SET permissions=$1 WHERE role='manager'", b); e != nil {
			a.dbError(w, e)
			return
		}
	}
	var u models.User
	e := a.Repo.DB.QueryRow(r.Context(), "SELECT id,name,email,role,permissions FROM users WHERE role='manager'").Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Permissions)
	if e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, u)
}
func (a *API) stats(w http.ResponseWriter, r *http.Request) {
	u := middleware.User(r)
	if u.Role == "customer" {
		fail(w, 403, "Permission denied")
		return
	}
	out := map[string]int{}
	var n int
	for _, status := range []string{"available", "reserved", "sold"} {
		if service.Can(u, "vehicles") {
			if e := a.Repo.DB.QueryRow(r.Context(), "SELECT count(*) FROM vehicles WHERE data->>'status'=$1", status).Scan(&n); e != nil {
				a.dbError(w, e)
				return
			}
			out[status] = n
		}
	}
	if service.Can(u, "chat") {
		a.Repo.DB.QueryRow(r.Context(), "SELECT count(*) FROM conversations WHERE status='open'").Scan(&n)
		out["conversations"] = n
	}
	JSON(w, 200, out)
}
func (a *API) enquiry(w http.ResponseWriter, r *http.Request) {
	if a.limited(w, r, "enquiry", 5) {
		return
	}
	var b struct {
		VehicleID string `json:"vehicle_id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Body      string `json:"body"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !service.UUID(b.VehicleID) || len(strings.TrimSpace(b.Name)) < 2 || len(b.Name) > 100 || !service.Email(b.Email) || len(b.Phone) > 30 || len(strings.TrimSpace(b.Body)) < 3 || len(b.Body) > 4000 {
		fail(w, 400, "Check enquiry fields")
		return
	}
	var exists bool
	a.Repo.DB.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM vehicles WHERE id=$1 AND (data->>'published')::boolean)", b.VehicleID).Scan(&exists)
	if !exists {
		fail(w, 404, "Vehicle not found")
		return
	}
	_, e := a.Repo.DB.Exec(r.Context(), "INSERT INTO enquiries(vehicle_id,name,email,phone,body) VALUES($1,$2,$3,$4,$5)", b.VehicleID, b.Name, b.Email, b.Phone, b.Body)
	if e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 201, map[string]bool{"ok": true})
}
func (a *API) enquiries(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		var b struct {
			Status string `json:"status"`
		}
		if !decode(w, r, &b) {
			return
		}
		if !service.UUID(r.PathValue("id")) || (b.Status != "new" && b.Status != "contacted" && b.Status != "closed") {
			fail(w, 400, "Invalid enquiry status")
			return
		}
		tag, e := a.Repo.DB.Exec(r.Context(), "UPDATE enquiries SET status=$1 WHERE id=$2", b.Status, r.PathValue("id"))
		if e != nil {
			a.dbError(w, e)
			return
		}
		if tag.RowsAffected() == 0 {
			fail(w, 404, "Enquiry not found")
			return
		}
		JSON(w, 200, b)
		return
	}
	rows, e := a.Repo.DB.Query(r.Context(), "SELECT row_to_json(e) FROM (SELECT e.*,v.data->>'title' AS vehicle_title FROM enquiries e LEFT JOIN vehicles v ON v.id=e.vehicle_id ORDER BY e.created_at DESC LIMIT 500)e")
	if e != nil {
		a.dbError(w, e)
		return
	}
	defer rows.Close()
	out := []json.RawMessage{}
	for rows.Next() {
		var b json.RawMessage
		if e = rows.Scan(&b); e != nil {
			a.dbError(w, e)
			return
		}
		out = append(out, b)
	}
	if e = rows.Err(); e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, out)
}
func (a *API) upload(w http.ResponseWriter, r *http.Request) {
	u := middleware.User(r)
	if !service.Can(u, "vehicles") && u.Role != "owner" {
		fail(w, 403, "Permission denied")
		return
	}
	if a.limited(w, r, "upload", 20) {
		return
	}
	if e := r.ParseMultipartForm(10 << 20); e != nil {
		fail(w, 400, "Image must be under 10 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
	f, _, e := r.FormFile("image")
	if e != nil {
		fail(w, 400, "Select an image")
		return
	}
	defer f.Close()
	tmp, e := os.CreateTemp("", "vehicle-*")
	if e != nil {
		a.dbError(w, e)
		return
	}
	defer os.Remove(tmp.Name())
	_, e = io.Copy(tmp, io.LimitReader(f, (10<<20)+1))
	tmp.Close()
	if e != nil {
		fail(w, 400, "Unable to read image")
		return
	}
	name := service.Token() + ".webp"
	out := filepath.Join(a.Uploads, name)
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	if e = exec.CommandContext(ctx, "python3", a.Optimizer, tmp.Name(), out).Run(); e != nil {
		os.Remove(out)
		fail(w, 400, "Use a valid JPEG, PNG or WebP image, up to 10 MB and 24 megapixels")
		return
	}
	JSON(w, 201, map[string]string{"url": "/uploads/" + name})
}

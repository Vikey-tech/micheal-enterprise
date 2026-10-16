package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"image"
	"image/color"
	"image/png"
	"io"
	"michealenterprise/internal/repository"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	repo, e := repository.Open(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	defer repo.DB.Close()
	if e = repo.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	// Run only against a disposable test database: these tables are cleared.
	if _, e = repo.DB.Exec(ctx, "TRUNCATE messages,conversations,enquiries,sessions,vehicles,users CASCADE"); e != nil {
		t.Fatal(e)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password-long"), 4)
	for _, role := range []string{"owner", "manager"} {
		_, e = repo.DB.Exec(ctx, "INSERT INTO users(name,email,password_hash,role) VALUES($1,$2,$3,$1)", role, role+"@example.com", string(hash))
		if e != nil {
			t.Fatal(e)
		}
	}
	optimizer, _ := filepath.Abs("../../../tools/optimize_images.py")
	api := New(repo, "http://test.local", t.TempDir(), optimizer, false)
	server := httptest.NewServer(api.Routes())
	defer server.Close()
	request := func(method, path string, body any, cookie *http.Cookie, want int) []byte {
		t.Helper()
		var data io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			data = bytes.NewReader(b)
		}
		r, _ := http.NewRequest(method, server.URL+path, data)
		r.Header.Set("Origin", "http://test.local")
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		resp, e := http.DefaultClient.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != want {
			t.Fatalf("%s %s: want %d got %d %s", method, path, want, resp.StatusCode, b)
		}
		return b
	}
	login := func(email string) *http.Cookie {
		t.Helper()
		b, _ := json.Marshal(map[string]string{"email": email, "password": "test-password-long"})
		r, _ := http.NewRequest("POST", server.URL+"/api/auth/login", bytes.NewReader(b))
		r.Header.Set("Origin", "http://test.local")
		resp, e := http.DefaultClient.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("login %s: %d", email, resp.StatusCode)
		}
		c := resp.Cookies()[0]
		if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
			t.Fatal("insecure cookie")
		}
		return c
	}
	owner, manager := login("owner@example.com"), login("manager@example.com")
	request("POST", "/api/auth/register", map[string]string{"name": "Customer", "email": "customer@example.com", "password": "test-password-long", "role": "owner"}, nil, 400)
	for _, email := range []string{"customer@example.com", "other@example.com"} {
		request("POST", "/api/auth/register", map[string]string{"name": "Customer", "email": email, "password": "test-password-long"}, nil, 200)
	}
	customer, other := login("customer@example.com"), login("other@example.com")
	request("POST", "/api/admin/vehicles", map[string]any{}, customer, 403)
	request("PUT", "/api/admin/manager", map[string]bool{"chat": true}, manager, 403)
	vehicle := map[string]any{"title": "Toyota Camry", "make": "Toyota", "model": "Camry", "year": 2022, "price": 24000000, "mileage": 12000, "fuel": "Petrol", "transmission": "Automatic", "status": "available", "published": false, "featured": true, "images": []string{}}
	var v map[string]any
	json.Unmarshal(request("POST", "/api/admin/vehicles", vehicle, manager, 200), &v)
	id := v["id"].(string)
	request("GET", "/api/vehicles/"+id, nil, nil, 404)
	vehicle["published"] = true
	request("PUT", "/api/admin/vehicles/"+id, vehicle, manager, 200)
	request("GET", "/api/vehicles/"+id, nil, nil, 200)
	request("POST", "/api/enquiries", map[string]string{"vehicle_id": id, "name": "Customer", "email": "customer@example.com", "phone": "12345", "body": "Can I arrange an inspection?"}, nil, 201)
	request("GET", "/api/admin/enquiries", nil, customer, 403)
	request("GET", "/api/admin/enquiries", nil, owner, 200)
	var conversation map[string]string
	json.Unmarshal(request("POST", "/api/conversations", map[string]string{}, customer, 200), &conversation)
	cid := conversation["id"]
	request("GET", "/api/conversations/"+cid+"/messages", nil, other, 404)
	header := http.Header{"Origin": []string{"http://test.local"}, "Cookie": []string{owner.String()}}
	conn, _, e := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http:", "ws:", 1)+"/api/ws", header)
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()
	request("POST", "/api/conversations/"+cid+"/messages", map[string]string{"body": "Is this still available?"}, customer, 201)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var event map[string]any
	if e = conn.ReadJSON(&event); e != nil {
		t.Fatal(e)
	}
	if event["body"] != "Is this still available?" {
		t.Fatal("websocket delivery failed")
	}
	request("POST", "/api/conversations/"+cid+"/messages", map[string]string{"body": "Yes, let us arrange a viewing."}, manager, 201)
	var messages []any
	json.Unmarshal(request("GET", "/api/conversations/"+cid+"/messages", nil, customer, 200), &messages)
	if len(messages) != 2 {
		t.Fatal("messages not persisted")
	}
	request("PUT", "/api/admin/manager", map[string]bool{"vehicles": false, "chat": false, "enquiries": false}, owner, 200)
	request("PUT", "/api/admin/vehicles/"+id, vehicle, manager, 403)
	request("GET", "/api/conversations/"+cid+"/messages", nil, manager, 404)
	upload := func(content []byte, want int) map[string]string {
		t.Helper()
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		file, _ := mw.CreateFormFile("image", "vehicle.png")
		file.Write(content)
		mw.Close()
		r, _ := http.NewRequest("POST", server.URL+"/api/admin/uploads", &buf)
		r.Header.Set("Origin", "http://test.local")
		r.Header.Set("Content-Type", mw.FormDataContentType())
		r.AddCookie(owner)
		resp, e := http.DefaultClient.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != want {
			t.Fatalf("upload want %d got %d: %s", want, resp.StatusCode, b)
		}
		var out map[string]string
		json.Unmarshal(b, &out)
		return out
	}
	upload([]byte("not a photo"), 400)
	photo := image.NewRGBA(image.Rect(0, 0, 20, 20))
	photo.Set(1, 1, color.White)
	var data bytes.Buffer
	png.Encode(&data, photo)
	urlImage := upload(data.Bytes(), 201)["url"]
	if !strings.HasSuffix(urlImage, ".webp") {
		t.Fatal("optimizer did not convert image")
	}
	request("GET", urlImage, nil, nil, 200)
	request("POST", "/api/auth/logout", map[string]string{}, customer, 200)
	request("GET", "/api/auth/me", nil, customer, 401)
}

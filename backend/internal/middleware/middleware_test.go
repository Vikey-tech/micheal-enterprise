package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOriginEnforcement(t *testing.T) {
	h := Security("https://cars.example", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for _, tc := range []struct {
		method, origin string
		want           int
	}{{"POST", "https://evil.example", 403}, {"POST", "", 403}, {"POST", "https://cars.example", 204}, {"GET", "", 204}} {
		r := httptest.NewRequest(tc.method, "/", nil)
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%+v got %d", tc, w.Code)
		}
	}
}
func TestLimiter(t *testing.T) {
	l := NewLimiter()
	if !l.Allow("a", 1) || l.Allow("a", 1) || !l.Allow("b", 1) {
		t.Fatal("rate limiting failed")
	}
}

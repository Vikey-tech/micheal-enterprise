package handlers

import (
	"context"
	"github.com/gorilla/websocket"
	"michealenterprise/internal/middleware"
	"michealenterprise/internal/models"
	"michealenterprise/internal/service"
	"net/http"
	"strings"
	"time"
)

func (a *API) chatAllowed(w http.ResponseWriter, r *http.Request) bool {
	u := middleware.User(r)
	if u.Role != "customer" && !service.Can(u, "chat") {
		fail(w, 403, "Chat permission denied")
		return false
	}
	return true
}
func (a *API) conversations(w http.ResponseWriter, r *http.Request) {
	if !a.chatAllowed(w, r) {
		return
	}
	u := middleware.User(r)
	if r.Method == "POST" {
		if u.Role != "customer" {
			fail(w, 403, "Only customers can start conversations")
			return
		}
		var id string
		e := a.Repo.DB.QueryRow(r.Context(), "INSERT INTO conversations(customer_id) VALUES($1) ON CONFLICT(customer_id) DO UPDATE SET customer_id=EXCLUDED.customer_id RETURNING id", u.ID).Scan(&id)
		if e != nil {
			a.dbError(w, e)
			return
		}
		JSON(w, 200, map[string]string{"id": id})
		return
	}
	v, e := a.Repo.Conversations(r.Context(), u)
	if e != nil {
		a.dbError(w, e)
		return
	}
	JSON(w, 200, v)
}
func (a *API) conversationStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !service.UUID(id) {
		fail(w, 400, "Invalid conversation")
		return
	}
	var b struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &b) {
		return
	}
	if b.Status != "open" && b.Status != "closed" {
		fail(w, 400, "Invalid status")
		return
	}
	tag, e := a.Repo.DB.Exec(r.Context(), "UPDATE conversations SET status=$1 WHERE id=$2", b.Status, id)
	if e != nil {
		a.dbError(w, e)
		return
	}
	if tag.RowsAffected() == 0 {
		fail(w, 404, "Conversation not found")
		return
	}
	JSON(w, 200, b)
}
func (a *API) messages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	u := middleware.User(r)
	if !service.UUID(id) || !a.Repo.CanAccessConversation(r.Context(), u, id) {
		fail(w, 404, "Conversation not found")
		return
	}
	if r.Method == "GET" {
		v, e := a.Repo.Messages(r.Context(), id)
		if e != nil {
			a.dbError(w, e)
			return
		}
		JSON(w, 200, v)
		return
	}
	if !a.limit.Allow("message:"+u.ID, 40) {
		fail(w, 429, "Please slow down")
		return
	}
	var b struct {
		Body string `json:"body"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Body = strings.TrimSpace(b.Body)
	if len(b.Body) < 1 || len(b.Body) > 4000 {
		fail(w, 400, "Message must be 1–4000 bytes")
		return
	}
	m, e := a.Repo.Send(r.Context(), u, id, b.Body)
	if e != nil {
		a.dbError(w, e)
		return
	}
	a.notify(m)
	JSON(w, 201, m)
}

type subscriber struct {
	user   models.User
	events chan models.Message
}

func (a *API) notify(m models.Message) {
	a.sockets.Range(func(k, v any) bool {
		s := v.(*subscriber)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if a.Repo.CanAccessConversation(ctx, s.user, m.ConversationID) {
			select {
			case s.events <- m:
			default:
			}
		}
		return true
	})
}
func (a *API) ws(w http.ResponseWriter, r *http.Request) {
	if !a.chatAllowed(w, r) {
		return
	}
	u := middleware.User(r)
	if !a.limit.Allow("ws:"+u.ID, 10) {
		fail(w, 429, "Too many connections")
		return
	}
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == a.Origin }, HandshakeTimeout: 5 * time.Second}
	c, e := up.Upgrade(w, r, nil)
	if e != nil {
		return
	}
	defer c.Close()
	s := &subscriber{u, make(chan models.Message, 32)}
	a.sockets.Store(c, s)
	defer a.sockets.Delete(c)
	done := make(chan struct{})
	c.SetReadLimit(1024)
	c.SetReadDeadline(time.Now().Add(70 * time.Second))
	c.SetPongHandler(func(string) error { return c.SetReadDeadline(time.Now().Add(70 * time.Second)) })
	go func() {
		defer close(done)
		for {
			if _, _, e := c.ReadMessage(); e != nil {
				return
			}
		}
	}()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	cookie, _ := r.Cookie("me_session")
	for {
		select {
		case <-done:
			return
		case <-r.Context().Done():
			return
		case m := <-s.events:
			current, e := a.Repo.Session(r.Context(), cookie.Value)
			if e != nil || !a.Repo.CanAccessConversation(r.Context(), current, m.ConversationID) {
				return
			}
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if c.WriteJSON(m) != nil {
				return
			}
		case <-ticker.C:
			if _, e := a.Repo.Session(r.Context(), cookie.Value); e != nil {
				return
			}
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if c.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		}
	}
}

package repository

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"michealenterprise/internal/models"
	"time"
)

//go:embed schema.sql
var schema string

type Repository struct{ DB *pgxpool.Pool }

func Open(ctx context.Context, url string) (*Repository, error) {
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return &Repository{db}, nil
}
func (s *Repository) Migrate(ctx context.Context) error { _, err := s.DB.Exec(ctx, schema); return err }
func HashToken(t string) string                         { v := sha256.Sum256([]byte(t)); return hex.EncodeToString(v[:]) }
func (s *Repository) UserByEmail(ctx context.Context, email string) (models.User, string, error) {
	var u models.User
	var hash string
	err := s.DB.QueryRow(ctx, "SELECT id,name,email,role,permissions,password_hash FROM users WHERE email=$1", email).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Permissions, &hash)
	return u, hash, err
}
func (s *Repository) Session(ctx context.Context, token string) (models.User, error) {
	var u models.User
	err := s.DB.QueryRow(ctx, "SELECT u.id,u.name,u.email,u.role,u.permissions FROM users u JOIN sessions s ON s.user_id=u.id WHERE s.token_hash=$1 AND s.expires_at>now()", HashToken(token)).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Permissions)
	return u, err
}
func (s *Repository) NewSession(ctx context.Context, id, token string) error {
	_, err := s.DB.Exec(ctx, "INSERT INTO sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)", HashToken(token), id, time.Now().Add(24*time.Hour))
	return err
}
func (s *Repository) Vehicles(ctx context.Context, admin bool) ([]models.Vehicle, error) {
	rows, err := s.DB.Query(ctx, "SELECT id,data FROM vehicles WHERE $1 OR (data->>'published')::boolean=true ORDER BY created_at DESC LIMIT 1000", admin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Vehicle{}
	for rows.Next() {
		var id string
		var b []byte
		if err = rows.Scan(&id, &b); err != nil {
			return nil, err
		}
		var v models.Vehicle
		if err = json.Unmarshal(b, &v); err != nil {
			return nil, err
		}
		v.ID = id
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Repository) Profile(ctx context.Context) (models.Profile, error) {
	var p models.Profile
	var b []byte
	err := s.DB.QueryRow(ctx, "SELECT data FROM profile WHERE id=1").Scan(&b)
	if err == nil {
		err = json.Unmarshal(b, &p)
	}
	return p, err
}
func (s *Repository) Conversations(ctx context.Context, u models.User) ([]models.Conversation, error) {
	rows, err := s.DB.Query(ctx, "SELECT c.id,c.customer_id,u.name,c.status,c.updated_at FROM conversations c JOIN users u ON u.id=c.customer_id WHERE $1 OR c.customer_id=$2::uuid ORDER BY c.updated_at DESC LIMIT 500", u.Role != "customer", u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Conversation{}
	for rows.Next() {
		var c models.Conversation
		if err = rows.Scan(&c.ID, &c.CustomerID, &c.CustomerName, &c.Status, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *Repository) CanAccessConversation(ctx context.Context, u models.User, id string) bool {
	var ok bool
	err := s.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM conversations WHERE id=$1::uuid AND ($2 OR customer_id=$3::uuid))", id, u.Role != "customer" && (u.Role == "owner" || u.Permissions["chat"]), u.ID).Scan(&ok)
	return err == nil && ok
}
func (s *Repository) Messages(ctx context.Context, id string) ([]models.Message, error) {
	rows, err := s.DB.Query(ctx, "SELECT m.id,m.conversation_id,m.sender_id,u.name,u.role,m.body,m.created_at FROM (SELECT * FROM messages WHERE conversation_id=$1 ORDER BY created_at DESC LIMIT 500) m JOIN users u ON u.id=m.sender_id ORDER BY m.created_at", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Message{}
	for rows.Next() {
		var m models.Message
		if err = rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.SenderName, &m.Role, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Repository) Send(ctx context.Context, u models.User, id, body string) (models.Message, error) {
	var m models.Message
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return m, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, "INSERT INTO messages(conversation_id,sender_id,body) VALUES($1,$2,$3) RETURNING id,created_at", id, u.ID, body).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	_, err = tx.Exec(ctx, "UPDATE conversations SET updated_at=now(),status='open' WHERE id=$1", id)
	if err != nil {
		return m, err
	}
	err = tx.Commit(ctx)
	m.ConversationID = id
	m.SenderID = u.ID
	m.SenderName = u.Name
	m.Role = u.Role
	m.Body = body
	return m, err
}

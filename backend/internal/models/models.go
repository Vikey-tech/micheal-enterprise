package models

import "time"

type User struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Role        string          `json:"role"`
	Permissions map[string]bool `json:"permissions"`
}
type Vehicle struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Make         string   `json:"make"`
	Model        string   `json:"model"`
	Year         int      `json:"year"`
	Price        int64    `json:"price"`
	Mileage      int      `json:"mileage"`
	Fuel         string   `json:"fuel"`
	Transmission string   `json:"transmission"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Published    bool     `json:"published"`
	Featured     bool     `json:"featured"`
	Images       []string `json:"images"`
}
type Profile struct {
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Phone    string `json:"phone"`
	WhatsApp string `json:"whatsapp"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Image    string `json:"image"`
}
type Conversation struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Status       string    `json:"status"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	SenderName     string    `json:"sender_name"`
	Role           string    `json:"role"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

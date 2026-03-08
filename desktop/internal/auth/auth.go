package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"desktop/internal/db"
)

type Session struct {
	Username  string
	Token     string
	CreatedAt time.Time
}

type Manager struct {
	db       *db.DB
	mu       sync.RWMutex
	sessions map[string]*Session // token -> session
}

func NewManager(database *db.DB) *Manager {
	return &Manager{db: database, sessions: make(map[string]*Session)}
}

func (m *Manager) Register(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = m.db.Exec(
		"INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, ?)",
		username, string(hash), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (m *Manager) Login(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}
	var hash string
	err := m.db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid username or password")
	}
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	m.mu.Lock()
	m.sessions[token] = &Session{Username: username, Token: token, CreatedAt: time.Now()}
	m.mu.Unlock()

	m.db.Exec("UPDATE users SET last_login = ? WHERE username = ?",
		time.Now().UTC().Format(time.RFC3339), username)

	return token, nil
}

func (m *Manager) Verify(token string) (string, error) {
	m.mu.RLock()
	session, ok := m.sessions[token]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("invalid or expired token")
	}
	return session.Username, nil
}

func (m *Manager) Logout(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

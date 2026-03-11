package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	DB *pgxpool.Pool
}

var sessions = make(map[string]string)
var mu sync.Mutex

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler
func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = app.DB.Exec(context.Background(),
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		creds.Username, string(hashedPassword))

	if err != nil {
		http.Error(w, "Username already taken or database error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully!"))
}

// LoginHandler
func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var expectedPasswordHash string
	err = app.DB.QueryRow(context.Background(),
		"SELECT password_hash FROM users WHERE username = $1",
		creds.Username).Scan(&expectedPasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(expectedPasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionToken := generateSessionToken()

	mu.Lock()
	sessions[sessionToken] = creds.Username
	mu.Unlock()

	// ENTERPRISE COOKIE SETTINGS FOR CROSS-DOMAIN (Vercel + Render)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour), // 24-hour session
		HttpOnly: true,
		Secure:   true,                  // REQUIRED for SameSite=None
		SameSite: http.SameSiteNoneMode, // REQUIRED for cross-domain cookies
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged in!"))
}

// LogoutHandler
func (app *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		w.WriteHeader(http.StatusOK) // Already logged out
		return
	}

	mu.Lock()
	delete(sessions, cookie.Value)
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// ProtectedHandler (Now returns JSON!)
func (app *App) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	username, exists := sessions[cookie.Value]
	mu.Unlock()

	if !exists {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// PROFESSIONAL JSON RESPONSE
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
		"status":   "authenticated",
	})
}

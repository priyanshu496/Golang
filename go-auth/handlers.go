package main

import (
	"context" // Needed for database timeouts and cancellations
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5" // The enterprise Postgres driver
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- Enterprise Dependency Injection ---

// App holds our application state. By attaching our database pool to this struct,
// all our handlers can access the database safely.
type App struct {
	DB *pgxpool.Pool
}

// --- In-Memory Session Store (To be moved to Redis later) ---

var sessions = make(map[string]string)
var mu sync.Mutex

// --- Helper Functions ---

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// --- HTTP Handlers (Now methods on the App struct!) ---

// RegisterHandler handles the creation of a new user using PostgreSQL
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

	// ENTERPRISE SQL: Insert the new user into the database
	// app.DB.Exec runs a query that doesn't return rows (like INSERT, UPDATE, DELETE)
	// The $1 and $2 are safe placeholders that prevent SQL Injection attacks!
	_, err = app.DB.Exec(context.Background(), 
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)", 
		creds.Username, string(hashedPassword))

	if err != nil {
		// A real enterprise app would check specifically for a unique constraint error here,
		// but for now, we will assume an error means the username is taken.
		http.Error(w, "Username already taken or database error", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully!\n"))
}

// LoginHandler verifies credentials against PostgreSQL and sets a cookie
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

	// ENTERPRISE SQL: Fetch the user's password hash
	// app.DB.QueryRow fetches exactly ONE row. We then Scan() the result into our variable.
	err = app.DB.QueryRow(context.Background(), 
		"SELECT password_hash FROM users WHERE username = $1", 
		creds.Username).Scan(&expectedPasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			// pgx.ErrNoRows specifically means no user was found with that username
			http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Compare the database hash with the typed password
	err = bcrypt.CompareHashAndPassword([]byte(expectedPasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionToken := generateSessionToken()

	mu.Lock()
	sessions[sessionToken] = creds.Username
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged in!\n"))
}

// LogoutHandler and ProtectedHandler also become methods on the App struct
func (app *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Unauthorized: No session found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value

	mu.Lock()
	delete(sessions, sessionToken)
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully logged out!\n"))
}

func (app *App) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized: Please log in", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	username, exists := sessions[cookie.Value]
	mu.Unlock()

	if !exists {
		http.Error(w, "Unauthorized: Invalid or expired session", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the secret dashboard, " + username + "!\n"))
}
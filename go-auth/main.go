package main

import (
	"fmt"
	"log"
	"net/http"
	"os" // Added to read environment variables!

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/priyanshu496/Golang.git/docs"
	"github.com/priyanshu496/Golang.git/database" 
)

func main() {
	// 1. Load the secret variables from the .env file (for local dev)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found or error loading it (expected in production)")
	}

	// 2. Initialize the Database Connection Pool
	pool, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer pool.Close() 

	// 3. Create our App instance
	app := &App{
		DB: pool,
	}

	// 4. Register our routes
	http.HandleFunc("/register", enableCORS(app.RegisterHandler))
	http.HandleFunc("/login", enableCORS(app.LoginHandler))
	http.HandleFunc("/logout", enableCORS(app.LogoutHandler))
	http.HandleFunc("/protected", enableCORS(app.ProtectedHandler))

	// Swagger route
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// ENTERPRISE FIX: Render dynamically assigns a port. We cannot hardcode 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local testing
	}

	fmt.Printf("Server starting on port %s\n", port)
	
	// Start the server
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

// --- Middleware ---

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ENTERPRISE FIX: Read the allowed frontend URL from environment variables
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000" // Fallback for local testing
		}

		// 1. Dynamically allow requests from Vercel (or localhost)
		w.Header().Set("Access-Control-Allow-Origin", frontendURL)
		
		// 2. Allow cookies to be sent back and forth
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 3. Allow these specific HTTP methods and headers
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 4. Handle "Preflight" requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 5. Pass to the actual handler
		next(w, r)
	}
}
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/priyanshu496/Golang.git/docs"
	"github.com/priyanshu496/Golang.git/database" // Importing our enterprise DB package!
)

func main() {
	// 1. Load the secret variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found or error loading it")
	}

	// 2. Initialize the Database Connection Pool
	pool, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	// Defer ensures the database pool closes cleanly when the server eventually shuts down
	defer pool.Close() 

	// 3. Create our App instance (Dependency Injection in action!)
	app := &App{
		DB: pool,
	}

	// 4. Register our routes using the app's newly attached methods
	// We wrap each handler in our new enableCORS middleware!
	http.HandleFunc("/register", enableCORS(app.RegisterHandler))
	http.HandleFunc("/login", enableCORS(app.LoginHandler))
	http.HandleFunc("/logout", enableCORS(app.LogoutHandler))
	http.HandleFunc("/protected", enableCORS(app.ProtectedHandler))

	// Swagger route (No need for CORS here since we view it directly in the browser)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	
	// Start the server
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

// --- Middleware ---

// enableCORS is a middleware function that adds necessary headers to allow Next.js to talk to Go
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Allow requests from our Next.js frontend (running on port 3000)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		
		// 2. Allow cookies to be sent back and forth! (Crucial for our sessions)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 3. Allow these specific HTTP methods and headers
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 4. Handle "Preflight" requests. Browsers send an automatic OPTIONS request first
		// to check if they are allowed to send the real POST/GET request.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 5. If everything is good, pass the request on to the actual handler (like LoginHandler)
		next(w, r)
	}
}
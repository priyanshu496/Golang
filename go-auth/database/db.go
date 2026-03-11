package database

import (
	"context"
	"fmt"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDB initializes our enterprise connection pool
func InitDB() (*pgxpool.Pool, error) {
	// 1. Get the URL from the environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// 2. Create the connection pool
	// context.Background() is used because this pool needs to stay alive
	// for the entire lifetime of the server.
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// 3. Ping the database to ensure our connection actually works
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database did not respond to ping: %v", err)
	}

	fmt.Println("Successfully connected to Neon PostgreSQL!")
	return pool, nil
}
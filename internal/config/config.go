package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/joho/godotenv"
)

type APIConfig struct {
	DB       *database.Queries
	PolyKey  string
	AlphaKey string
}

func Load() *APIConfig {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Warning: .env file not found!")
	}

	// Initialize database
	dbURL := mustGetEnv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get environment variables
	polyKey := mustGetEnv("POLYGON_API")
	alphaKey := mustGetEnv("ALPHA_API")

	dbQueries := database.New(db)
	cfg := &APIConfig{
		DB:       dbQueries,
		PolyKey:  polyKey,
		AlphaKey: alphaKey,
	}
	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s not set", key)
	}
	return val
}

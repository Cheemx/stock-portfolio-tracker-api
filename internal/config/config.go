package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type APIConfig struct {
	DB        *database.Queries
	RD        *redis.Client
	JWTSecret string
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
	// defer db.Close()

	// Connect with redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Adding a redis consumer group
	err = rdb.XGroupCreateMkStream(context.Background(), "events:liveStocks", "events-group", "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Error creating consumer group: %v", err)
	}
	fmt.Println("Redis Stream Created successfully!")

	dbQueries := database.New(db)
	cfg := &APIConfig{
		DB:        dbQueries,
		RD:        rdb,
		JWTSecret: mustGetEnv("JWT_SECRET"),
	}
	fmt.Println("Redis Client Connected Successfully.")
	fmt.Println("Postgres Database Connected Successfully.")
	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s not set", key)
	}
	return val
}

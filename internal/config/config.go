package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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

var slidingWindowScript = redis.NewScript(`
local key = KEYS[1] -- lua indexing start from 1!
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

-- Removing old requests outside the window!
redis.call("ZREMRANGEBYSCORE", key, 0, now - window)

-- count remaining requests
local count = redis.call("ZCARD", key) -- ZCARD returns cardinality(number of elements) of a sorted set stored at given key.

if count >= limit then
	return 0
end

-- Limit not crossed so add it to set of request-timestamps
-- key: name of the sorted set -> quota:<ipAddr>:<action>
-- now: score -> double precision floating point number representing score of the member
-- now: member -> Unique string identifier (what else can be a better unique identifier than current Time)
redis.call("ZADD", key, now, now)

-- Keep the key alive at least for timeWindow duration
redis.call("EXPIRE", key, window)

return 1
`)

func (cfg *APIConfig) CheckRateLimit(ctx context.Context, ipAddr, action string) bool {
	var timeWindowInSeconds int64
	var limit int64

	switch action {
	case "transaction":
		timeWindowInSeconds = 60
		limit = 10
	case "signup":
		timeWindowInSeconds = 600
		limit = 5
	case "login":
		timeWindowInSeconds = 600
		limit = 5
	default:
		timeWindowInSeconds = 3600
		limit = 100
	}

	key := fmt.Sprintf("quota:%s:%s", ipAddr, action)
	now := time.Now().Unix()

	allowed, err := slidingWindowScript.Run(ctx, cfg.RD, []string{key}, now, timeWindowInSeconds, limit).Bool()

	if err != nil {
		log.Printf("Rate limit check failed: %v", err)
		return false
	}

	return allowed
}

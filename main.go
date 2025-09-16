package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("Can't get DB_URL from .env")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	var now string
	err = db.QueryRow("SELECT NOW()").Scan(&now)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB connected, current time:", now)
	fmt.Println("Hello Stocks")
}

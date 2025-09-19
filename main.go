package main

import (
	"log"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/routes"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/worker"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const port = "8080"

func main() {
	r := gin.Default()

	gin.SetMode(gin.DebugMode)

	cfg := config.Load()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Stocker Panicked: %v\n", r)
			}
		}()
		worker.Stocker(cfg)
	}()
	go worker.ProcessStocks(cfg)

	routes.UserRoutes(r, cfg)
	routes.TransactionRoutes(r, cfg)
	log.Printf("Serving Stock tracker API on port: %s\n", port)
	log.Fatal(r.Run(":" + port))
}

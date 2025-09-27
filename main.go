package main

import (
	"log"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/events"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/routes"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/worker"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const port = "8080"

func main() {
	r := gin.Default()

	gin.SetMode(gin.DebugMode)

	cfg := config.Load()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Backend is working!"})
	})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://cheems-writes.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Stocker Panicked: %v\n", r)
			}
		}()
		worker.Stocker(cfg)
	}()
	go worker.ProcessStocks(cfg)
	go events.HubInstance.Run()

	routes.UserRoutes(r, cfg)
	routes.TransactionRoutes(r, cfg)
	routes.HoldingRoutes(r, cfg)
	routes.PortfolioRoutes(r, cfg)
	routes.StockRoutes(r, cfg)
	routes.SSERoutes(r, cfg)
	log.Printf("Serving Stock tracker API on port: %s\n", port)
	log.Fatal(r.Run(":" + port))
}

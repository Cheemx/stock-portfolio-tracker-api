package main

import (
	"log"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/routes"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const port = "8080"

func main() {
	r := gin.Default()

	cfg := config.Load()

	routes.UserRoutes(r, cfg)
	log.Printf("Serving Stock tracker API on port: %s\n", port)
	log.Fatal(r.Run(":" + port))
}

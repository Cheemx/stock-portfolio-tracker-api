package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func SSERoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.GET("/api/events", controllers.HandleSSE(cfg))
}

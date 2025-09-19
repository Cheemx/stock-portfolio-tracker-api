package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func HoldingRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.GET("/api/holdings", controllers.Holdings(cfg))
}

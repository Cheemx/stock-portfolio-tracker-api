package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func PortfolioRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.GET("/api/portfolio", controllers.Portfolio(cfg))
}

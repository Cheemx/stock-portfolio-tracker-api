package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func StockRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.GET("/api/stocks", controllers.GetStocks(cfg))
	router.GET("/api/stocks/search", controllers.SearchStocks(cfg))
}

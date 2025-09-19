package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func TransactionRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.POST("/api/transactions", controllers.CreateTransaction(cfg))
	router.GET("/api/transactions", controllers.GetTransactions(cfg))
}

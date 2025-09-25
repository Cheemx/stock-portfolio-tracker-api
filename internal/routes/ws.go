package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func WsRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.GET("/ws", controllers.HandleWS(cfg))
}

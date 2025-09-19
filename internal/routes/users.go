package routes

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine, cfg *config.APIConfig) {
	router.POST("/api/users", controllers.CreateUser(cfg))
	router.POST("/admin/reset", controllers.DeleteUsers(cfg))
	router.POST("/api/login", controllers.LoginUser(cfg))
}

package controllers

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
)

func CreateUser(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req database.CreateUserParams

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(401, gin.H{
				"error": "invalid request payload",
			})
			return
		}

		// Create user in database
		user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
			Email: req.Email,
			Name:  req.Name,
		})
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "failed to create user",
			})
		}

		// Respond with user data
		ctx.JSON(201, user)
	}
}

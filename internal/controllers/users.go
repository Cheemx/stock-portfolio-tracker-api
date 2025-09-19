package controllers

import (
	"log"
	"os"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateUser(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(401, gin.H{
				"error": "invalid request payload",
			})
			return
		}

		// hashing the text password from req
		hashedPass, err := auth.HashPassword(req.Password)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "Error in auth package hashing password",
			})
			return
		}

		// Create user in database
		user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
			Email:          req.Email,
			Name:           req.Name,
			HashedPassword: hashedPass,
		})
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Creating response
		res := struct {
			ID        uuid.UUID `json:"id"`
			Name      string    `json:"name"`
			CreatedAt time.Time `json:"created_at"`
			Email     string    `json:"email"`
		}{
			ID:        user.ID,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			Email:     user.Email,
		}

		// Respond with user data
		ctx.JSON(201, res)
	}
}

func LoginUser(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// request parsing
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(401, gin.H{
				"error": "invalid request payload",
			})
			return
		}

		// Get User by Email
		user, err := cfg.DB.GetUserByEmail(ctx, req.Email)
		if err != nil {
			ctx.JSON(401, gin.H{
				"error": "user with this email not found",
			})
			return
		}

		// validating password
		err = auth.CheckPasswordHash(req.Password, user.HashedPassword)
		if err != nil {
			ctx.JSON(401, gin.H{
				"error": "email password not matching!",
			})
			return
		}

		// Create the Access Token
		token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, 3600*time.Second)
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": "Error making JWT",
			})
			return
		}

		// Creating response and responding
		res := struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			Email     string    `json:"email"`
			Name      string    `json:"name"`
			Token     string    `json:"token"`
		}{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			Email:     user.Email,
			Name:      user.Name,
			Token:     token,
		}

		ctx.JSON(200, res)
	}
}

func DeleteUsers(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if os.Getenv("PLATFORM") != "dev" {
			ctx.Writer.WriteHeader(403)
			return
		}

		err := cfg.DB.DeleteAllUsers(ctx)
		if err != nil {
			log.Printf("Error Deleting all Users: %v", err)
			ctx.Writer.WriteHeader(500)
			return
		}
		log.Println("All Users Deleted Successfully.")

		ctx.Writer.WriteHeader(200)
	}
}

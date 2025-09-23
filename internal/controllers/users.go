package controllers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateUser(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Applying rate limit on signup route
		if !cfg.CheckRateLimit(ctx, ctx.ClientIP(), "signup") {
			respondWithError(ctx, http.StatusTooManyRequests, "Sign-Up Quota Expired", nil)
			return
		}

		// request handling
		req := struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			respondWithError(ctx, 401, "error unmarshalling request", err)
			return
		}

		// hashing the text password from req
		hashedPass, err := auth.HashPassword(req.Password)
		if err != nil {
			respondWithError(ctx, 500, "error hashing password", err)
			return
		}

		// Create user in database
		user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
			Email:          req.Email,
			Name:           req.Name,
			HashedPassword: hashedPass,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_email_key\"") {
				respondWithError(ctx, 401, "User already exists", err)
				return
			}
			respondWithError(ctx, 500, "error creating user", err)
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
		// Applying rate limit on signup route
		if !cfg.CheckRateLimit(ctx, ctx.ClientIP(), "login") {
			respondWithError(ctx, http.StatusTooManyRequests, "Sign-Up Quota Expired", nil)
			return
		}

		// request parsing
		req := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			respondWithError(ctx, 401, "error unmarshalling request", err)
			return
		}

		// Get User by Email
		user, err := cfg.DB.GetUserByEmail(ctx, req.Email)
		if err != nil {
			respondWithError(ctx, 401, "user with email not found", err)
			return
		}

		// validating password
		err = auth.CheckPasswordHash(req.Password, user.HashedPassword)
		if err != nil {
			respondWithError(ctx, 401, "email and password doesn't match", err)
			return
		}

		// Create the Access Token
		token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, 3600*time.Second)
		if err != nil {
			respondWithError(ctx, 500, "error making token", err)
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
			respondWithError(ctx, 500, "error deleting all users", err)
			return
		}
		log.Println("All Users Deleted Successfully.")

		ctx.Writer.WriteHeader(200)
	}
}

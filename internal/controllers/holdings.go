package controllers

import (
	"net/http"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
)

func Holdings(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Applying rate limiter to limit viewing holdings
		if !cfg.CheckRateLimit(ctx, ctx.ClientIP(), "holdings") {
			respondWithError(ctx, http.StatusTooManyRequests, "Wait for some time!", nil)
			return
		}

		// Authorization required for this route
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "Authentication error", err)
		}

		res, err := GetHoldings(ctx, cfg, userId)
		if err != nil {
			respondWithError(ctx, 500, "holdings not found for this user", err)
			return
		}

		// return holdings
		ctx.JSON(200, res)
	}
}

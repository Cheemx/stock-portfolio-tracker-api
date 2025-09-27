package controllers

import (
	"database/sql"
	"net/http"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
)

func Portfolio(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Applying rate limiter to limit requesting portfolio
		if !cfg.CheckRateLimit(ctx, ctx.ClientIP(), "portfolio") {
			respondWithError(ctx, http.StatusTooManyRequests, "Wait for some time!", nil)
			return
		}
		// Authorization required for this route
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "Authentication error", err)
			return
		}

		res, err := GetPortfolio(ctx, cfg, userId)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(ctx, 404, "NO Portfoilio for this user", err)
				return
			}
			respondWithError(ctx, 500, "Error getting the portfolio for current user", err)
			return
		}

		ctx.JSON(200, res)
	}
}

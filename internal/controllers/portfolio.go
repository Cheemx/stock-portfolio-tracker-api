package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

		// Construct the portfolio response
		type PortfolioRes struct {
			TotalInvested     float64 `json:"total_invested"`
			CurrentValue      float64 `json:"current_value"`
			TotalProfitOrLoss float64 `json:"pnl"`
			PNLPercentage     float64 `json:"pnl_percentage"`
			HoldingsCount     int     `json:"holdings_count"`
		}

		// Get the portfolio from cache if present OR not expired!
		var portfolioRes PortfolioRes
		portfolioBytes, err := cfg.RD.Get(ctx, "portfolio:"+userId.String()).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(portfolioBytes), &portfolioRes); err == nil {
				ctx.JSON(200, portfolioRes)
				return
			}
		} else if err != redis.Nil {
			log.Printf("Redis GET error: %v\n", err)
		}

		// get the portfolio for the user
		portfolio, err := cfg.DB.GetPortfolioForUser(ctx, userId)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(ctx, 404, "NO Portfoilio for this user", err)
				return
			}
			respondWithError(ctx, 500, "Error getting the portfolio for current user", err)
			return
		}

		// Add the pnl and pnlpercentage
		pnlPercentage := 0.0
		pnl := portfolio.CurrentValue - portfolio.TotalInvested
		if portfolio.TotalInvested > 0 {
			pnlPercentage = (pnl / portfolio.TotalInvested) * 100
		}

		// respond with the portfolio
		res := PortfolioRes{
			TotalInvested:     portfolio.TotalInvested,
			CurrentValue:      portfolio.CurrentValue,
			TotalProfitOrLoss: pnl,
			PNLPercentage:     pnlPercentage,
			HoldingsCount:     int(portfolio.HoldingsCount),
		}
		resJSON, err := json.Marshal(res)
		if err != nil {
			log.Printf("Error marshalling portfolio cache: %v\n", err)
		}

		// Set the portfolio Result in json
		err = cfg.RD.Set(ctx, "portfolio:"+userId.String(), resJSON, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Error setting portfolio cache: %v\n", err)
		}

		ctx.JSON(200, res)
	}
}

package controllers

import (
	"database/sql"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
)

func Portfolio(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Authorization required for this route
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "Authentication error", err)
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
		pnl := portfolio.CurrentValue - portfolio.TotalInvested
		pnlPercentage := (pnl / portfolio.TotalInvested) * 100

		// Construct the portfolio response
		type PortfolioRes struct {
			TotalInvested     float64 `json:"total_invested"`
			CurrentValue      float64 `json:"current_value"`
			TotalProfitOrLoss float64 `json:"pnl"`
			PNLPercentage     float64 `json:"pnl_percentage"`
			HoldingsCount     int     `json:"holdings_count"`
		}

		// respond with the portfolio
		res := PortfolioRes{
			TotalInvested:     float64(portfolio.TotalInvested),
			CurrentValue:      float64(portfolio.CurrentValue),
			TotalProfitOrLoss: float64(pnl),
			PNLPercentage:     float64(pnlPercentage),
			HoldingsCount:     int(portfolio.HoldingsCount),
		}

		ctx.JSON(200, res)
	}
}

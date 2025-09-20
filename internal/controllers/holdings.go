package controllers

import (
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
)

func Holdings(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Authorization required for this route
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "Authentication error", err)
		}

		// Get holdings from user
		holdings, err := cfg.DB.GetAllHoldingsForUser(ctx, userId)
		if err != nil {
			respondWithError(ctx, 500, "holdings not found for this user", err)
			return
		}

		// Draft response struct with pnl and pnlPercentage
		type holdingRes struct {
			StockSymbol            string  `json:"stock_symbol"`
			CompanyName            string  `json:"company_name"`
			Quantity               int     `json:"quantity"`
			AveragePrice           float64 `json:"average_price"`
			CurrentPrice           float64 `json:"curr_price"`
			CurrentValue           float64 `json:"curr_evaluation"`
			ProfitOrLoss           float64 `json:"pnl"`
			ProfitOrLossPercentage float64 `json:"pnl_percentage"`
			TotalInvested          float64 `json:"total_invested"`
		}
		var res []holdingRes

		// calculate pnl and pnlpercentage for each holding and store in res
		for _, holding := range holdings {
			currValue := float64(holding.Quantity) * holding.CurrentPrice
			pnl := currValue - holding.TotalInvested
			pnlPercentage := (pnl / holding.TotalInvested) * 100
			hold := holdingRes{
				StockSymbol:            holding.StockSymbol,
				CompanyName:            holding.CompanyName,
				Quantity:               int(holding.Quantity),
				AveragePrice:           holding.AveragePrice,
				CurrentPrice:           holding.CurrentPrice,
				CurrentValue:           currValue,
				ProfitOrLoss:           pnl,
				ProfitOrLossPercentage: pnlPercentage,
				TotalInvested:          holding.TotalInvested,
			}

			res = append(res, hold)
		}
		// return holdings
		ctx.JSON(200, res)
	}
}

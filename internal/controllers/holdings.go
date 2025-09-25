package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

		// Get the holdings from cache if present OR not expired
		var holdingsRes []holdingRes
		holdingsBytes, err := cfg.RD.Get(ctx, "holdings:"+userId.String()).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(holdingsBytes), &holdingsRes); err == nil {
				ctx.JSON(200, holdingsRes)
				return
			}
		} else if err != redis.Nil {
			log.Printf("Redis GET error: %v\n", err)
		}

		// Get holdings from user
		holdings, err := cfg.DB.GetAllHoldingsForUser(ctx, userId)
		if err != nil {
			respondWithError(ctx, 500, "holdings not found for this user", err)
			return
		}

		// calculate pnl and pnlpercentage for each holding and store in res
		var res []holdingRes
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

		// Set the holdings in Cache for faster lookup
		// This cache holdings will be deleted in the transactions.go whenever
		// A new transaction is successful
		resJSON, err := json.Marshal(res)
		if err != nil {
			log.Printf("Error marshalling holdings cache: %v\n", err)
		}

		err = cfg.RD.Set(ctx, "holdings:"+userId.String(), resJSON, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Error setting holdings cache: %v\n", err)
		}

		// return holdings
		ctx.JSON(200, res)
	}
}

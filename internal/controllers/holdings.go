package controllers

import (
	"strconv"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Holdings(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// authenticate user first and get auth header
		token, err := auth.GetBearerToken(ctx.Request.Header)
		if err != nil {
			respondWithError(ctx, 401, "token not present", err)
			return
		}

		// get userId from user
		userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "token and user doesn't match", err)
			return
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
		}
		holdingsLength := len(holdings)
		res := make([]holdingRes, holdingsLength)

		// calculate pnl and pnlpercentage for each holding and store in res
		for _, holding := range holdings {
			totalInvested, err := CalculateTotalInvestedForAStock(cfg, ctx, userId, holding)
			if err != nil {
				respondWithError(ctx, 500, "error calculating pnl", err)
				return
			}
			avgPrice, _ := strconv.ParseFloat(holding.AveragePrice, 64)
			currPrice, _ := strconv.ParseFloat(holding.CurrentPrice, 64)
			currValue := float64(holding.Quantity) * currPrice
			pnl := currValue - totalInvested
			pnlPercentage := (pnl / totalInvested) * 100
			hold := holdingRes{
				StockSymbol:            holding.StockSymbol,
				CompanyName:            holding.CompanyName,
				Quantity:               int(holding.Quantity),
				AveragePrice:           avgPrice,
				CurrentPrice:           currPrice,
				CurrentValue:           currValue,
				ProfitOrLoss:           pnl,
				ProfitOrLossPercentage: pnlPercentage,
			}

			res = append(res, hold)
		}
		// return holdings
		ctx.JSON(200, res)
	}
}

func CalculateTotalInvestedForAStock(cfg *config.APIConfig, ctx *gin.Context, userId uuid.UUID, holding database.GetAllHoldingsForUserRow) (float64, error) {
	txns, err := cfg.DB.GetAllTransactionsForUserBySymbol(ctx, database.GetAllTransactionsForUserBySymbolParams{
		UserID:      userId,
		StockSymbol: holding.StockSymbol,
	})
	if err != nil {
		return 0.0, err
	}

	totalInvested := 0.0
	for _, txn := range txns {
		totalAmount, _ := strconv.ParseFloat(txn.TotalAmount, 64)
		if txn.Type == buy {
			totalInvested += totalAmount
		}
		if txn.Type == sell {
			totalInvested -= totalAmount
		}
	}
	return totalInvested, nil
}

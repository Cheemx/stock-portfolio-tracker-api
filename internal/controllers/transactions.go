package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/utils"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/worker"
	"github.com/gin-gonic/gin"
)

const (
	buy  = "BUY"
	sell = "SELL"
)

func CreateTransaction(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Authenticated route niga!
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, http.StatusUnauthorized, "Authentication error", err)
			return
		}

		// Parse request
		var req struct {
			StockSymbol string `json:"stock_symbol"`
			Type        string `json:"type"`
			Quantity    int    `json:"quantity"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			respondWithError(ctx, http.StatusBadRequest, "Invalid request body", err)
			return
		}
		if req.Quantity <= 0 || (req.Type != buy && req.Type != sell) {
			respondWithError(ctx, http.StatusBadRequest, "Quantity must be > 0 and type must be BUY/SELL", nil)
			return
		}

		// Get stock info
		stonk, err := getOrFetchStock(ctx, cfg, req.StockSymbol)
		if err != nil {
			respondWithError(ctx, http.StatusInternalServerError, "Failed to resolve stock info", err)
			return
		}

		// Get current holdings for user
		currHolding, err := cfg.DB.GetHoldingByStockSymbol(ctx, database.GetHoldingByStockSymbolParams{
			UserID:      userId,
			StockSymbol: req.StockSymbol,
		})
		isNewHolding := errors.Is(err, sql.ErrNoRows)
		if err != nil && !isNewHolding {
			respondWithError(ctx, http.StatusInternalServerError, "Error fetching holdings", err)
			return
		}

		// Compute transaction outcome
		var (
			newQuantity int
			newAvg      float64
			totalAmount float64
		)
		if isNewHolding {
			// Cannot sell if no holdings
			if req.Type == sell {
				respondWithError(ctx, http.StatusBadRequest, "Can't sell the stock you don't OWN niga!", nil)
				return
			}
			newQuantity, _, newAvg, _, _, totalAmount =
				utils.HandleBuyTransaction(req.Quantity, 0, 0, stonk.CurrentPrice)
		} else {
			switch req.Type {
			case buy:
				newQuantity, _, newAvg, _, _, totalAmount =
					utils.HandleBuyTransaction(req.Quantity, int(currHolding.Quantity), currHolding.AveragePrice, stonk.CurrentPrice)
				currHolding.TotalInvested += totalAmount
			case sell:
				newQuantity, _, newAvg, _, _, totalAmount =
					utils.HandleSellTransaction(req.Quantity, int(currHolding.Quantity), currHolding.AveragePrice, stonk.CurrentPrice)
				currHolding.TotalInvested -= totalAmount
			}
		}

		// Insert transaction record
		txn, err := cfg.DB.CreateATransaction(ctx, database.CreateATransactionParams{
			UserID:      userId,
			StockSymbol: req.StockSymbol,
			Type:        req.Type,
			Quantity:    int32(req.Quantity),
			Price:       stonk.CurrentPrice,
			TotalAmount: totalAmount,
		})
		if err != nil {
			respondWithError(ctx, http.StatusInternalServerError, "Failed to create transaction", err)
			return
		}

		// Update or remove holding
		if newQuantity == 0 && !isNewHolding {
			if _, err := cfg.DB.DeleteHoldingsOnSellOut(ctx, database.DeleteHoldingsOnSellOutParams{
				UserID:      userId,
				StockSymbol: req.StockSymbol,
			}); err != nil {
				respondWithError(ctx, http.StatusInternalServerError, "Failed to delete holding", err)
				return
			}
			ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Sold out holdings for %s", req.StockSymbol)})
			return
		}

		updatedHolding, err := cfg.DB.CreateNewHoldingOrUpdateExistingForUser(ctx,
			database.CreateNewHoldingOrUpdateExistingForUserParams{
				UserID:        userId,
				StockSymbol:   req.StockSymbol,
				Quantity:      int32(newQuantity),
				AveragePrice:  newAvg,
				TotalInvested: currHolding.TotalInvested + totalAmount,
			})
		if err != nil {
			respondWithError(ctx, http.StatusInternalServerError, "Failed to update holdings", err)
			return
		}

		// Respond with Transaction and Current HOlding
		ctx.JSON(http.StatusCreated, gin.H{
			"Transaction": txn,
			"Holding":     updatedHolding,
		})
	}
}

// Helper to get stock from DB or Yahoo
func getOrFetchStock(ctx context.Context, cfg *config.APIConfig, symbol string) (database.Stock, error) {
	stonk, err := cfg.DB.GetStockBySymbol(ctx, symbol)
	if err == nil {
		return stonk, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return database.Stock{}, err
	}

	// Fetch from Yahoo if not in DB
	var client http.Client
	stonkFromYahoo, err := worker.FetchFromYahoo(symbol, &client)
	if err != nil {
		return database.Stock{}, err
	}

	return cfg.DB.CreateNewStockOrUpdateExisting(ctx, database.CreateNewStockOrUpdateExistingParams{
		Symbol:        stonkFromYahoo.Symbol,
		CompanyName:   stonkFromYahoo.CompanyName,
		CurrentPrice:  stonkFromYahoo.CurrentPrice,
		PreviousClose: stonkFromYahoo.PreviousClose,
	})
}

func GetTransactions(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Authorization required for this route
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "Authentication error", err)
		}

		// get transactions for userId
		txns, err := cfg.DB.GetAllTransactionsForUser(ctx, userId)
		if err != nil {
			respondWithError(ctx, 500, "error getting transactions", err)
			return
		}

		// return the transactions
		ctx.JSON(200, txns)
	}
}

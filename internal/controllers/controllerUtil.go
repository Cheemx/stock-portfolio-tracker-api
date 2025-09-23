package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/worker"
	"github.com/gin-gonic/gin"
)

const (
	buy  = "BUY"
	sell = "SELL"
)

func respondWithError(ctx *gin.Context, statusCode int, errorString string, err error) {
	ctx.JSON(statusCode, gin.H{
		"Error": fmt.Sprintf("%s: %v\n", errorString, err),
	})
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

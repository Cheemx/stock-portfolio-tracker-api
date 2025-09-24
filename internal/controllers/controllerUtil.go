package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

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
	// Fetch from cache
	stockJSON, err := cfg.RD.Get(ctx, "stock:"+symbol).Result()
	if err == nil {
		var stockRes database.Stock
		if err := json.Unmarshal([]byte(stockJSON), &stockRes); err == nil {
			return stockRes, nil
		}
	}

	// Fetch from DB if not in cache!
	stonk, err := cfg.DB.GetStockBySymbol(ctx, symbol)
	if err == nil {
		// Warm the cache
		if stockJSON, err := json.Marshal(stonk); err == nil {
			cfg.RD.Set(ctx, "stock:"+stonk.Symbol, stockJSON, 30*time.Second)
		}
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

	created, err := cfg.DB.CreateNewStockOrUpdateExisting(ctx, database.CreateNewStockOrUpdateExistingParams{
		Symbol:        stonkFromYahoo.Symbol,
		CompanyName:   stonkFromYahoo.CompanyName,
		CurrentPrice:  stonkFromYahoo.CurrentPrice,
		PreviousClose: stonkFromYahoo.PreviousClose,
	})
	if err != nil {
		return database.Stock{}, err
	}

	// warm the cache after creating
	if stockJSON, err := json.Marshal(created); err == nil {
		cfg.RD.Set(ctx, "stock:"+created.Symbol, stockJSON, 30*time.Second)
	}

	return created, err
}

package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
)

const (
	buy      = "BUY"
	sell     = "SELL"
	YahooAPI = "https://query1.finance.yahoo.com/v8/finance/chart/"
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
	stonkFromYahoo, err := FetchFromYahoo(symbol, &client)
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

// Util to fetch stock data from free YahooAPI
func FetchFromYahoo(symbol string, client *http.Client) (database.Stock, error) {
	var resp config.YahooFinanceResponse
	reqToStockAPI, err := http.NewRequest("GET", YahooAPI+symbol, nil)
	if err != nil {
		return database.Stock{}, err
	}

	// Adding headers to avoid 429 from YahooAPI
	reqToStockAPI.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	reqToStockAPI.Header.Set("Accept", "application/json")

	yahooRes, err := client.Do(reqToStockAPI)
	if err != nil {
		return database.Stock{}, err
	}
	defer yahooRes.Body.Close()
	if yahooRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(yahooRes.Body)
		return database.Stock{}, fmt.Errorf("yahoo api error: %s - %s", yahooRes.Status, string(body))
	}
	if err := json.NewDecoder(yahooRes.Body).Decode(&resp); err != nil {
		return database.Stock{}, err
	}

	if len(resp.Chart.Result) == 0 {
		return database.Stock{}, fmt.Errorf("no results in Yahoo response")
	}
	yahooResult := resp.Chart.Result[0]
	return yahooResult.ToStock(), nil
}

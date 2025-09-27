package worker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/controllers"
	"github.com/redis/go-redis/v9"
)

func Stocker(cfg *config.APIConfig) {
	// DB call to get all symbol for current user
	// symbols is list of symbols for holdings owned all userbase
	symbols, err := cfg.DB.GetStockSymbolsOfHoldings(context.Background())
	if err != nil {
		log.Printf("Error getting symbols from DB: %v\n", err)
	}

	if len(symbols) < 5 {
		symbols = append(symbols, []string{"AAPL", "MSFT", "RELIANCE.NS", "TCS.NS", "HDFCBANK.NS", "^NSEI"}...)
	}

	thirtySecTicker := time.NewTicker(30 * time.Second)
	defer thirtySecTicker.Stop()

	for range thirtySecTicker.C {
		now := time.Now()

		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			continue
		}

		hour := now.Hour()
		if hour < 9 || hour > 16 {
			continue
		}
		client := &http.Client{Timeout: 5 * time.Second}
		for _, symbol := range symbols {
			// Fetching stock from Yahoo API
			stockRes, err := controllers.FetchFromYahoo(symbol, client)
			if err != nil {
				log.Printf("error fetching from YahooAPI: %v\n", err)
				continue
			}

			// Pushing stockJSON ([]byte) in redis Stream
			stockJSON, _ := json.Marshal(stockRes)
			err = cfg.RD.XAdd(context.Background(), &redis.XAddArgs{
				Stream: "events:liveStocks",
				Values: map[string]any{"stock": string(stockJSON)},
				MaxLen: 100,
				Approx: true,
			}).Err()
			if err != nil {
				log.Printf("error adding stock in redis pipeline: %v\n", err)
				continue
			}
		}
	}
}

package worker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/redis/go-redis/v9"
)

const YahooAPI = "https://query1.finance.yahoo.com/v8/finance/chart/"

func Stocker(cfg *config.APIConfig) {
	var yahooResult config.YahooResult

	// DB call to get all symbol for current user
	// symbols is list of symbols for holdings owned all userbase
	symbols, err := cfg.DB.GetStockSymbolsOfHoldings(context.Background())
	if err != nil {
		log.Printf("Error getting symbols from DB: %v\n", err)
	}

	if len(symbols) < 1 {
		log.Println("Niga first buy some stonks to hold them")
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
		client := &http.Client{}
		for _, symbol := range symbols {
			reqToStockAPI, err := http.NewRequest("GET", YahooAPI+symbol, nil)
			if err != nil {
				log.Printf("Error creating HTTP request to API: %v\n", err)
				continue
			}

			yahooRes, err := client.Do(reqToStockAPI)
			if err != nil {
				log.Printf("Error making Free API Call: %v\n", err)
				continue
			}
			if err := json.NewDecoder(yahooRes.Body).Decode(&yahooResult); err != nil {
				log.Print("Error decoding response from free API in YahooResult\n")
				continue
			}
			yahooJSON, _ := json.Marshal(yahooResult)
			err = cfg.RD.XAdd(context.Background(), &redis.XAddArgs{
				Stream: "events:liveStocks",
				Values: map[string]any{"stock": string(yahooJSON)},
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

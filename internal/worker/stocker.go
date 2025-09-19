package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/redis/go-redis/v9"
)

const YahooAPI = "https://query1.finance.yahoo.com/v8/finance/chart/"

func Stocker(cfg *config.APIConfig) {
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
			stockRes, err := FetchFromYahoo(symbol, client)
			if err != nil {
				log.Printf("error fetching from YahooAPI: %v\n", err)
				continue
			}
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

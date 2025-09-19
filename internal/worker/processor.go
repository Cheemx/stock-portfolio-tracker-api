package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/redis/go-redis/v9"
)

func ProcessStocks(cfg *config.APIConfig) {
	for {
		// Reading from Redis Streams
		result, err := cfg.RD.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    "events-group",
			Consumer: "consumer-dbFlusher",
			Streams:  []string{"events:liveStocks", ">"},
			Block:    0,
		}).Result()

		if err != nil {
			log.Printf("Error reading from stream: %v\n", err)
			time.Sleep(15 * time.Second)
			continue
		}

		for _, stream := range result {
			for _, message := range stream.Messages {
				stockJSON := message.Values["stock"].(string)
				var stockRes database.Stock
				err := json.Unmarshal([]byte(stockJSON), &stockRes)
				if err != nil {
					log.Printf("Error unmarshaling event: %v\n", err)
					continue
				}

				// store in Postgres DB
				_, err = cfg.DB.CreateNewStockOrUpdateExisting(context.Background(), database.CreateNewStockOrUpdateExistingParams{
					Symbol:        stockRes.Symbol,
					CompanyName:   stockRes.CompanyName,
					CurrentPrice:  stockRes.CurrentPrice,
					PreviousClose: stockRes.PreviousClose,
				})

				if err != nil {
					log.Printf("DB insert error: %v\n", err)
					continue
				}

				// Acknowledge processing of the stock
				err = cfg.RD.XAck(context.Background(), "events:liveStocks", "events-group", message.ID).Err()
				if err != nil {
					log.Printf("Error acknowledging the read in stream: %v\n", err)
					continue
				}
			}
		}

	}
}

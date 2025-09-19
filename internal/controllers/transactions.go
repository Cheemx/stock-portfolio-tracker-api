package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
		// Authorization required for this route
		token, err := auth.GetBearerToken(ctx.Request.Header)
		if err != nil {
			respondWithError(ctx, 401, "unauthorized access", err)
			return
		}

		// get UserID from token
		userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, 401, "token doesn't match", err)
			return
		}

		// expected request body
		req := struct {
			StockSymbol string `json:"stock_symbol"`
			Type        string `json:"type"`
			Quantity    int    `json:"quantity"`
		}{}

		// Bind JSON to req struct
		if err := ctx.ShouldBindJSON(&req); err != nil {
			respondWithError(ctx, http.StatusBadRequest, "error unmarshalling request", err)
			return
		}

		// Validating the request
		if req.Quantity <= 0 || (req.Type != buy && req.Type != sell) {
			respondWithError(ctx, http.StatusBadRequest, "You can only BUY or SELL positive stocks", nil)
			return
		}

		// get the stock info which user is buying
		var stonk database.Stock
		stonk, err = cfg.DB.GetStockBySymbol(ctx, req.StockSymbol)
		if err != nil {
			var client http.Client
			if err == sql.ErrNoRows {
				stonkFromYahoo, err := worker.FetchFromYahoo(req.StockSymbol, &client)
				if err != nil {
					respondWithError(ctx, 500, "error fetching stock data from API", err)
					return
				}
				stonk, err = cfg.DB.CreateNewStockOrUpdateExisting(ctx, database.CreateNewStockOrUpdateExistingParams{
					Symbol:        stonkFromYahoo.Symbol,
					CompanyName:   stonkFromYahoo.CompanyName,
					CurrentPrice:  stonkFromYahoo.CurrentPrice,
					PreviousClose: stonkFromYahoo.PreviousClose,
				})
				if err != nil {
					respondWithError(ctx, 500, "error upserting fresh stock data", err)
					return
				}
			} else {
				respondWithError(ctx, 500, "error getting stonk from database", err)
				return
			}
		}
		// parsing current stock price stored in db in float64
		currPrice, err := strconv.ParseFloat(stonk.CurrentPrice, 64)
		if err != nil {
			respondWithError(ctx, 500, "error parsing current Stock price", err)
			return
		}

		// get user's current holdings for given stock
		currHoldings, err := cfg.DB.GetHoldingByStockSymbol(ctx, database.GetHoldingByStockSymbolParams{
			UserID:      userId,
			StockSymbol: req.StockSymbol,
		})
		if err != nil && err != sql.ErrNoRows {
			respondWithError(ctx, 500, "error getting current holdings", err)
			return
		}

		// If new User or buying first time
		if err == sql.ErrNoRows {
			if req.Type == sell {
				respondWithError(ctx, http.StatusBadRequest, "You can't sell the stocks you don't own", nil)
				return
			}
			currQuant, _, currAvg, _, _, totalAmount := utils.HandleBuyTransaction(req.Quantity, 0, 0, currPrice)

			// Push transaction in DB
			log.Printf("Creating txn for user=%s, symbol=%s, price=%s", userId, req.StockSymbol, stonk.CurrentPrice)
			totalAmountStr := strconv.FormatFloat(totalAmount, 'f', 2, 64)
			txn, err := cfg.DB.CreateATransaction(ctx, database.CreateATransactionParams{
				UserID:      userId,
				StockSymbol: req.StockSymbol,
				Type:        buy,
				Quantity:    int32(req.Quantity),
				Price:       stonk.CurrentPrice,
				TotalAmount: totalAmountStr,
			})
			if err != nil {
				respondWithError(ctx, 500, "error flushing txn in DB", err)
				return
			}

			// Update the user's holdings based on this transaction
			currAvgPrice := strconv.FormatFloat(currAvg, 'f', 2, 64)
			holding, err := cfg.DB.CreateNewHoldingOrUpdateExistingForUser(ctx, database.CreateNewHoldingOrUpdateExistingForUserParams{
				UserID:       userId,
				StockSymbol:  txn.StockSymbol,
				Quantity:     int32(currQuant),
				AveragePrice: currAvgPrice,
			})
			if err != nil {
				respondWithError(ctx, 500, "error creating holding in DB", err)
				return
			}

			// Use totalInvested, pnl, pnlPercentage for portfolio as we're not using any table for portfolio then probably we'll think of something here
			// -- name: GetPortfolioForUser :one
			// SELECT
			//         SUM(holdings.quantity * holdings.average_price) AS total_invested,
			//         SUM(holdings.quantity * stocks.current_price) AS current_value,
			//         COUNT(holdings.user_id) AS holdings_count
			// FROM holdings
			// JOIN users
			// ON holdings.user_id = users.id
			// JOIN stocks
			// ON holdings.stock_symbol = stocks.symbol
			// WHERE users.id = $1
			// GROUP BY holdings.user_id;
			// Above is the kinda query to fetch portfolio! which will be implemented in it's controllernot here

			ctx.JSON(201, map[string]any{
				"Transaction": txn,
				"Holding":     holding,
			})
			return
		}

		// Parsing previous Average stored in DB in float64
		prevAvg, err := strconv.ParseFloat(currHoldings.AveragePrice, 64)
		if err != nil {
			respondWithError(ctx, 500, "error parsing current averagePrice", err)
			return
		}

		// getting info from utils based on transaction type
		var currQuant int
		var currAvg, totalAmount float64
		if req.Type == buy {
			currQuant, _, currAvg, _, _, totalAmount = utils.HandleBuyTransaction(req.Quantity, int(currHoldings.Quantity), prevAvg, currPrice)
		}
		if req.Type == sell {
			currQuant, _, currAvg, _, _, totalAmount = utils.HandleSellTransaction(req.Quantity, int(currHoldings.Quantity), prevAvg, currPrice)
		}

		// Push transaction in DB
		totalAmountStr := strconv.FormatFloat(totalAmount, 'f', 2, 64)
		txn, err := cfg.DB.CreateATransaction(ctx, database.CreateATransactionParams{
			UserID:      userId,
			StockSymbol: req.StockSymbol,
			Type:        req.Type,
			Quantity:    int32(req.Quantity),
			Price:       stonk.CurrentPrice,
			TotalAmount: totalAmountStr,
		})
		if err != nil {
			respondWithError(ctx, 500, "error flushing txn in DB", err)
			return
		}

		// If currQuant becomes zero delete the holding
		if currQuant == 0 {
			_, err := cfg.DB.DeleteHoldingsOnSellOut(ctx, database.DeleteHoldingsOnSellOutParams{
				UserID:      userId,
				StockSymbol: req.StockSymbol,
			})
			if err != nil {
				respondWithError(ctx, 500, "error removing holding on SellOut", err)
				return
			}
			ctx.JSON(201, fmt.Sprintf("Sold Out Holdings for %s", req.StockSymbol))
			return
		}

		// Update the user's holdings based on this transaction
		currAvgPrice := strconv.FormatFloat(currAvg, 'f', 2, 64)
		updatedHolding, err := cfg.DB.CreateNewHoldingOrUpdateExistingForUser(ctx, database.CreateNewHoldingOrUpdateExistingForUserParams{
			UserID:       userId,
			StockSymbol:  txn.StockSymbol,
			Quantity:     int32(currQuant),
			AveragePrice: currAvgPrice,
		})
		if err != nil {
			respondWithError(ctx, 500, "error updating holding in DB", err)
			return
		}

		ctx.JSON(201, map[string]any{
			"Transaction": txn,
			"Holding":     updatedHolding,
		})
	}
}

package controllers

import (
	"database/sql"
	"errors"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/gin-gonic/gin"
)

func GetStocks(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// just get the stocks from DB
		stonks, err := cfg.DB.GetAllStocks(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(ctx, 404, "No stocks found in DB", err)
				return
			}
			respondWithError(ctx, 500, "error getting stocks from DB", err)
			return
		}
		// return them
		ctx.JSON(200, stonks)
	}
}

func SearchStocks(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Bind query param to request
		req := struct {
			Query string `form:"q"`
		}{}
		err := ctx.ShouldBind(&req)
		if err != nil {
			respondWithError(ctx, 500, "error parsing query", err)
			return
		}

		// Get matching stock
		stonks, err := cfg.DB.SearchStockByName(ctx, sql.NullString{String: req.Query, Valid: true})
		if err != nil {
			respondWithError(ctx, 404, "No stocks found for this query", err)
			return
		}

		// Return matching stock
		ctx.JSON(200, stonks)
	}
}

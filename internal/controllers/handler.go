package controllers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/events"
	"github.com/gin-gonic/gin"
)

func HandleSSE(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// SSE headers
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")

		// Flush writer
		flusher, ok := ctx.Writer.(http.Flusher)
		if !ok {
			respondWithError(ctx, http.StatusInternalServerError, "streaming unsupported", nil)
			return
		}

		// Get userId from Authorization token
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, http.StatusUnauthorized, "Authorization token error", err)
			return
		}

		// Get Subscriptions(stock symbols) for userId
		symbols, err := cfg.DB.GetStockSymbolsForUser(ctx, userId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return
			}
			respondWithError(ctx, 500, "Error getting subscriptions for user", err)
			return
		}

		symbolSet := make(map[string]bool)
		for _, symbol := range symbols {
			symbolSet[symbol] = true
		}

		client := &events.Client{
			ID:      userId,
			Send:    make(chan []byte, 1024),
			Symbols: symbolSet,
		}

		events.HubInstance.Register <- client

		defer func() { events.HubInstance.Unregister <- client }()

		// Push updates until client closes connection
		notify := ctx.Writer.CloseNotify()

		for {
			select {
			case <-notify:
				return
			case msg, ok := <-client.Send:
				if !ok {
					return
				}
				_, _ = ctx.Writer.Write([]byte("data: "))
				_, _ = ctx.Writer.Write(msg)
				_, _ = ctx.Writer.Write([]byte("\n\n"))
				flusher.Flush()
			}
		}
	}
}

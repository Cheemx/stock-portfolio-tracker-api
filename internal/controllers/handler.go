package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/auth"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/config"
	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID   uuid.UUID
	Conn *websocket.Conn
	Send chan []byte
}

var Upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Minute,
	ReadBufferSize:   10 * 1024, // 10 KB for stock
	WriteBufferSize:  10 * 1024, // 10 KB for stock
	CheckOrigin: func(r *http.Request) bool {
		return true // For now let's accept everyone
	},
}

var (
	clients       = make(map[uuid.UUID]*Client)
	subscriptions = make(map[string]map[uuid.UUID]*Client) // symbol -> {userId -> *Client}
	register      = make(chan *Client)
	unregister    = make(chan *Client)
	Broadcast     = make(chan []byte, 10*1024)
)

func HandleWS(cfg *config.APIConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Upgrading to Websocket from HTTP
		conn, err := Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			respondWithError(ctx, http.StatusBadRequest, "Error upgrading from HTTP to Websocket", err)
			return
		}
		log.Println("Connection upgraded to Websocket successfully.")

		// Get userId from Authorization token
		userId, err := auth.GetUserID(ctx.Request.Header, cfg.JWTSecret)
		if err != nil {
			respondWithError(ctx, http.StatusUnauthorized, "Authorization token error", err)
			return
		}

		// Get Subscriptions(stock symbols of client) for userId
		symbols, err := cfg.DB.GetStockSymbolsForUser(ctx, userId)
		if err != nil {
			respondWithError(ctx, 500, "Error getting subscriptions for user", err)
			return
		}

		client := &Client{
			ID:   userId,
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		// Register client + their subscriptions
		register <- client
		for _, symbol := range symbols {
			if subscriptions[symbol] == nil {
				subscriptions[symbol] = make(map[uuid.UUID]*Client)
			}
			subscriptions[symbol][userId] = client
		}

		go client.writePump()
	}
}

func RunHub() {
	for {
		select {
		// Registering the client
		case client := <-register:
			clients[client.ID] = client

		// Unregistering and deleting the client + it's subscriptions
		case client := <-unregister:
			if _, ok := clients[client.ID]; ok {
				delete(clients, client.ID)
				for symbol := range subscriptions {
					delete(subscriptions[symbol], client.ID)
				}
				close(client.Send)
				client.Conn.Close()
			}

		// Grabbing stockJSON from Broadcast
		case stockJSON := <-Broadcast:
			var stock database.Stock
			err := json.Unmarshal(stockJSON, &stock)
			if err != nil {
				log.Printf("Error unmarshalling stockJSON: %v", err)
				continue
			}

			// Send updates to the clients who're subscribed to this symbol
			if subs, ok := subscriptions[stock.Symbol]; ok {
				for _, client := range subs {
					select {
					case client.Send <- stockJSON:
					default:
						close(client.Send)
						delete(clients, client.ID)
					}
				}
			}
		}
	}
}

func (c *Client) writePump() {
	for stockJSON := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, stockJSON); err != nil {
			break
		}
	}
}

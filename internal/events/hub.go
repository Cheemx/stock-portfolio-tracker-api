package events

import (
	"encoding/json"
	"log"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
	"github.com/google/uuid"
)

type Client struct {
	ID      uuid.UUID
	Send    chan []byte
	Symbols map[string]bool
}

type Hub struct {
	Clients    map[uuid.UUID]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

var HubInstance = &Hub{
	Clients:    make(map[uuid.UUID]*Client),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Broadcast:  make(chan []byte, 10*1024),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ID] = client

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}

		case stockJSON := <-h.Broadcast:
			var stock database.Stock
			if err := json.Unmarshal(stockJSON, &stock); err != nil {
				log.Printf("Error unmarshalling stockJSON: %v", err)
				continue
			}

			// fan-out only to clients who actually hold this stock
			for _, client := range h.Clients {
				if client.Symbols[stock.Symbol] {
					select {
					case client.Send <- stockJSON:
					default: // client too slow
						close(client.Send)
						delete(h.Clients, client.ID)
					}
				}
			}
		}
	}
}

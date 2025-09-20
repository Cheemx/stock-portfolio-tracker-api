package config

import (
	"database/sql"
	"time"

	"github.com/Cheemx/stock-portfolio-tacker-api/internal/database"
)

type YahooFinanceResponse struct {
	Chart YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooResult `json:"result"`
	Error  any           `json:"error"`
}

type YahooResult struct {
	Meta       YahooMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators YahooIndicators `json:"indicators"`
}

type YahooMeta struct {
	Currency            string  `json:"currency"`
	Symbol              string  `json:"symbol"`
	ExchangeName        string  `json:"exchangeName"`
	InstrumentType      string  `json:"instrumentType"`
	RegularMarketTime   int64   `json:"regularMarketTime"`
	RegularMarketPrice  float64 `json:"regularMarketPrice"`
	PreviousClose       float64 `json:"previousClose"`
	RegularMarketVolume int64   `json:"regularMarketVolume"`
	LongName            string  `json:"longName"`
	ShortName           string  `json:"shortName"`
}

type YahooIndicators struct {
	Quote []YahooQuote `json:"quote"`
}

type YahooQuote struct {
	High   []float64 `json:"high"`
	Open   []float64 `json:"open"`
	Low    []float64 `json:"low"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}

func (yr *YahooResult) ToStock() database.Stock {
	return database.Stock{
		Symbol:        yr.Meta.Symbol,
		CompanyName:   yr.Meta.LongName,
		CurrentPrice:  yr.Meta.RegularMarketPrice,
		PreviousClose: sql.NullFloat64{Float64: yr.Meta.PreviousClose, Valid: true},
		UpdatedAt:     time.Now(),
	}
}

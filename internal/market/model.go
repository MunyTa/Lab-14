package market

import "time"

type Tick struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type Candle struct {
	Symbol      string    `json:"symbol"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	Volume      float64   `json:"volume"`
	Count       int       `json:"count"`
	Source      string    `json:"source"`
}

package transport

import (
	"encoding/json"
	"io"
	"time"

	"github.com/MunyTa/Lab-14/internal/market"
)

type CandleRecordBatch struct {
	Symbols      []string    `json:"symbol"`
	WindowStarts []time.Time `json:"window_start"`
	WindowEnds   []time.Time `json:"window_end"`
	Opens        []float64   `json:"open"`
	Highs        []float64   `json:"high"`
	Lows         []float64   `json:"low"`
	Closes       []float64   `json:"close"`
	Volumes      []float64   `json:"volume"`
	Counts       []int       `json:"count"`
}

func NewCandleRecordBatch(candles []market.Candle) CandleRecordBatch {
	batch := CandleRecordBatch{
		Symbols:      make([]string, 0, len(candles)),
		WindowStarts: make([]time.Time, 0, len(candles)),
		WindowEnds:   make([]time.Time, 0, len(candles)),
		Opens:        make([]float64, 0, len(candles)),
		Highs:        make([]float64, 0, len(candles)),
		Lows:         make([]float64, 0, len(candles)),
		Closes:       make([]float64, 0, len(candles)),
		Volumes:      make([]float64, 0, len(candles)),
		Counts:       make([]int, 0, len(candles)),
	}

	for _, candle := range candles {
		batch.Symbols = append(batch.Symbols, candle.Symbol)
		batch.WindowStarts = append(batch.WindowStarts, candle.WindowStart)
		batch.WindowEnds = append(batch.WindowEnds, candle.WindowEnd)
		batch.Opens = append(batch.Opens, candle.Open)
		batch.Highs = append(batch.Highs, candle.High)
		batch.Lows = append(batch.Lows, candle.Low)
		batch.Closes = append(batch.Closes, candle.Close)
		batch.Volumes = append(batch.Volumes, candle.Volume)
		batch.Counts = append(batch.Counts, candle.Count)
	}
	return batch
}

func (b CandleRecordBatch) Rows() int {
	return len(b.Symbols)
}

func (b CandleRecordBatch) Schema() []string {
	return []string{"symbol", "window_start", "window_end", "open", "high", "low", "close", "volume", "count"}
}

func WriteRecordBatchJSON(writer io.Writer, candles []market.Candle) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(NewCandleRecordBatch(candles))
}


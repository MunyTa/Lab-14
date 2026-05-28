package transport

import (
	"testing"
	"time"

	"github.com/MunyTa/Lab-14/internal/market"
)

func TestCandleRecordBatchKeepsColumnarSchema(t *testing.T) {
	candles := []market.Candle{
		{
			Symbol:      "BTCUSDT",
			WindowStart: time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC),
			WindowEnd:   time.Date(2026, 5, 28, 12, 1, 0, 0, time.UTC),
			Open:        100,
			High:        110,
			Low:         95,
			Close:       104,
			Volume:      8,
			Count:       4,
		},
		{
			Symbol:      "ETHUSDT",
			WindowStart: time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC),
			WindowEnd:   time.Date(2026, 5, 28, 12, 1, 0, 0, time.UTC),
			Open:        10,
			High:        11,
			Low:         9,
			Close:       10.5,
			Volume:      14,
			Count:       5,
		},
	}

	batch := NewCandleRecordBatch(candles)
	if batch.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", batch.Rows())
	}
	if got, want := batch.Schema(), []string{"symbol", "window_start", "window_end", "open", "high", "low", "close", "volume", "count"}; !sameStrings(got, want) {
		t.Fatalf("unexpected schema: got %#v, want %#v", got, want)
	}
	if batch.Symbols[0] != "BTCUSDT" || batch.Closes[1] != 10.5 {
		t.Fatalf("batch lost candle values: %#v", batch)
	}
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}


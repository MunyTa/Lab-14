package market

import (
	"math"
	"testing"
	"time"
)

func TestWindowAggregatorBuildsOHLCVCandles(t *testing.T) {
	base := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	aggregator := NewWindowAggregator(2*time.Minute, 0)

	ticks := []Tick{
		{Symbol: "BTCUSDT", Price: 100, Quantity: 1.5, Timestamp: base.Add(10 * time.Second), Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 112, Quantity: 2, Timestamp: base.Add(40 * time.Second), Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 95, Quantity: 1, Timestamp: base.Add(80 * time.Second), Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 105, Quantity: 5.5, Timestamp: base.Add(110 * time.Second), Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 120, Quantity: 3, Timestamp: base.Add(130 * time.Second), Source: "simulator"},
	}

	var candles []Candle
	for _, tick := range ticks {
		candles = append(candles, aggregator.Add(tick)...)
	}
	candles = append(candles, aggregator.Flush()...)

	if len(candles) != 2 {
		t.Fatalf("expected 2 candles, got %d: %#v", len(candles), candles)
	}

	first := candles[0]
	if first.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected symbol: %s", first.Symbol)
	}
	if !first.WindowStart.Equal(base) || !first.WindowEnd.Equal(base.Add(2*time.Minute)) {
		t.Fatalf("unexpected window: %s - %s", first.WindowStart, first.WindowEnd)
	}
	assertFloat(t, "open", first.Open, 100)
	assertFloat(t, "high", first.High, 112)
	assertFloat(t, "low", first.Low, 95)
	assertFloat(t, "close", first.Close, 105)
	assertFloat(t, "volume", first.Volume, 10)
	if first.Count != 4 {
		t.Fatalf("expected 4 ticks in first candle, got %d", first.Count)
	}
}

func TestWindowAggregatorCanFlushByRecordLimit(t *testing.T) {
	base := time.Date(2026, 5, 28, 13, 0, 0, 0, time.UTC)
	aggregator := NewWindowAggregator(time.Minute, 3)

	var emitted []Candle
	for i, price := range []float64{10, 12, 11} {
		emitted = append(emitted, aggregator.Add(Tick{
			Symbol:    "ETHUSDT",
			Price:     price,
			Quantity:  1,
			Timestamp: base.Add(time.Duration(i) * time.Second),
			Source:    "simulator",
		})...)
	}

	if len(emitted) != 1 {
		t.Fatalf("expected record limit to emit one candle, got %d", len(emitted))
	}
	if emitted[0].Count != 3 {
		t.Fatalf("expected 3 ticks in emitted candle, got %d", emitted[0].Count)
	}
}

func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("%s: got %.8f, want %.8f", name, got, want)
	}
}


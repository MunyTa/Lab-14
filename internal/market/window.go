package market

import (
	"sort"
	"time"
)

type WindowAggregator struct {
	window      time.Duration
	recordLimit int
	current     map[string]*Candle
}

func NewWindowAggregator(window time.Duration, recordLimit int) *WindowAggregator {
	if window <= 0 {
		window = time.Minute
	}
	return &WindowAggregator{
		window:      window,
		recordLimit: recordLimit,
		current:     map[string]*Candle{},
	}
}

func (a *WindowAggregator) Add(tick Tick) []Candle {
	if err := ValidateTick(tick); err != nil {
		return nil
	}

	tick.Timestamp = tick.Timestamp.UTC()
	start := tick.Timestamp.Truncate(a.window)
	key := tick.Symbol

	active, ok := a.current[key]
	if ok && !active.WindowStart.Equal(start) {
		emitted := *active
		delete(a.current, key)
		a.current[key] = newCandle(tick, start, a.window)
		return []Candle{emitted}
	}
	if !ok {
		a.current[key] = newCandle(tick, start, a.window)
		return nil
	}

	updateCandle(active, tick)
	if a.recordLimit > 0 && active.Count >= a.recordLimit {
		emitted := *active
		delete(a.current, key)
		return []Candle{emitted}
	}
	return nil
}

func (a *WindowAggregator) Flush() []Candle {
	candles := make([]Candle, 0, len(a.current))
	for key, candle := range a.current {
		candles = append(candles, *candle)
		delete(a.current, key)
	}
	sort.Slice(candles, func(i, j int) bool {
		if candles[i].WindowStart.Equal(candles[j].WindowStart) {
			return candles[i].Symbol < candles[j].Symbol
		}
		return candles[i].WindowStart.Before(candles[j].WindowStart)
	})
	return candles
}

func newCandle(tick Tick, start time.Time, window time.Duration) *Candle {
	return &Candle{
		Symbol:      tick.Symbol,
		WindowStart: start,
		WindowEnd:   start.Add(window),
		Open:        tick.Price,
		High:        tick.Price,
		Low:         tick.Price,
		Close:       tick.Price,
		Volume:      tick.Quantity,
		Count:       1,
		Source:      tick.Source,
	}
}

func updateCandle(candle *Candle, tick Tick) {
	if tick.Price > candle.High {
		candle.High = tick.Price
	}
	if tick.Price < candle.Low {
		candle.Low = tick.Price
	}
	candle.Close = tick.Price
	candle.Volume += tick.Quantity
	candle.Count++
}


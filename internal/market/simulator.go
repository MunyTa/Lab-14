package market

import (
	"math"
	"time"
)

type Simulator struct {
	symbols []string
	now     time.Time
	step    int
}

func NewSimulator(symbols []string, start time.Time) *Simulator {
	if start.IsZero() {
		start = time.Now().UTC()
	}
	return &Simulator{symbols: symbols, now: start.UTC()}
}

func (s *Simulator) Next() []Tick {
	timestamp := s.now.Add(time.Duration(s.step) * time.Second)
	ticks := make([]Tick, 0, len(s.symbols))
	for index, symbol := range s.symbols {
		base := 1000.0 + float64(index+1)*250.0
		wave := math.Sin(float64(s.step+index)) * 12.5
		trend := float64(s.step%17) * 0.8
		ticks = append(ticks, Tick{
			Symbol:    symbol,
			Price:     round4(base + wave + trend),
			Quantity:  round4(0.25 + float64((s.step+index)%9)*0.15),
			Timestamp: timestamp,
			Source:    "simulator",
		})
	}
	s.step++
	return ticks
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

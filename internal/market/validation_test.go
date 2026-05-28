package market

import (
	"testing"
	"time"
)

func TestValidateTickRejectsInvalidMarketData(t *testing.T) {
	valid := Tick{
		Symbol:    "BTCUSDT",
		Price:     65000,
		Quantity:  0.25,
		Timestamp: time.Date(2026, 5, 28, 14, 0, 0, 0, time.UTC),
		Source:    "simulator",
	}
	if err := ValidateTick(valid); err != nil {
		t.Fatalf("valid tick rejected: %v", err)
	}

	cases := []Tick{
		{Symbol: "", Price: 65000, Quantity: 1, Timestamp: valid.Timestamp, Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 0, Quantity: 1, Timestamp: valid.Timestamp, Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 65000, Quantity: -1, Timestamp: valid.Timestamp, Source: "simulator"},
		{Symbol: "BTCUSDT", Price: 65000, Quantity: 1, Source: "simulator"},
	}

	for _, tick := range cases {
		if err := ValidateTick(tick); err == nil {
			t.Fatalf("expected invalid tick to be rejected: %#v", tick)
		}
	}
}


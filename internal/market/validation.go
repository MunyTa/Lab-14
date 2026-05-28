package market

import (
	"errors"
	"strings"
)

func ValidateTick(tick Tick) error {
	if strings.TrimSpace(tick.Symbol) == "" {
		return errors.New("symbol is required")
	}
	if tick.Price <= 0 {
		return errors.New("price must be positive")
	}
	if tick.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if tick.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	if strings.TrimSpace(tick.Source) == "" {
		return errors.New("source is required")
	}
	return nil
}


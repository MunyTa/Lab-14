//go:build !rustvalidate

package rustvalidator

import (
	"github.com/MunyTa/Lab-14/internal/market"
)

func Backend() string {
	return "go-fallback"
}

func ValidateTick(tick market.Tick) error {
	return market.ValidateTick(tick)
}

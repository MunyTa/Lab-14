//go:build rustvalidate

package rustvalidator

/*
#cgo LDFLAGS: -L${SRCDIR}/../../rust_validator/target/release -llab14_validator
#include <stdlib.h>

int validate_crypto_tick(const char* symbol, double price, double quantity, long long timestamp_unix);
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/MunyTa/Lab-14/internal/market"
)

func Backend() string {
	return "rust-cgo"
}

func ValidateTick(tick market.Tick) error {
	symbol := C.CString(tick.Symbol)
	defer C.free(unsafe.Pointer(symbol))

	code := C.validate_crypto_tick(
		symbol,
		C.double(tick.Price),
		C.double(tick.Quantity),
		C.longlong(tick.Timestamp.Unix()),
	)
	if code != 0 {
		return fmt.Errorf("rust validator rejected tick with code %d", int(code))
	}
	return nil
}

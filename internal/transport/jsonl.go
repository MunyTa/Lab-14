package transport

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MunyTa/Lab-14/internal/market"
)

func WriteCandlesJSONL(path string, candles []market.Candle) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buffered := bufio.NewWriter(file)
	defer buffered.Flush()

	encoder := json.NewEncoder(buffered)
	for _, candle := range candles {
		if err := encoder.Encode(candle); err != nil {
			return err
		}
	}
	return nil
}

func WriteBatchJSON(path string, candles []market.Candle) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteRecordBatchJSON(file, candles)
}


//go:build arrow

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/MunyTa/Lab-14/internal/market"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

func main() {
	input := flag.String("input", "out/candles.jsonl", "JSONL candles input")
	addr := flag.String("addr", ":8081", "HTTP listen address")
	flag.Parse()

	http.HandleFunc("/arrow", func(writer http.ResponseWriter, request *http.Request) {
		candles, err := readCandles(*input)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/vnd.apache.arrow.stream")
		if err := writeArrowIPC(writer, candles); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Printf("arrow server listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readCandles(path string) ([]market.Candle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var candles []market.Candle
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var candle market.Candle
		if err := json.Unmarshal(scanner.Bytes(), &candle); err != nil {
			return nil, err
		}
		candles = append(candles, candle)
	}
	return candles, scanner.Err()
}

func writeArrowIPC(writer http.ResponseWriter, candles []market.Candle) error {
	allocator := memory.NewGoAllocator()
	symbols := array.NewStringBuilder(allocator)
	starts := array.NewTimestampBuilder(allocator, &arrow.TimestampType{Unit: arrow.Microsecond, TimeZone: "UTC"})
	ends := array.NewTimestampBuilder(allocator, &arrow.TimestampType{Unit: arrow.Microsecond, TimeZone: "UTC"})
	opens := array.NewFloat64Builder(allocator)
	highs := array.NewFloat64Builder(allocator)
	lows := array.NewFloat64Builder(allocator)
	closes := array.NewFloat64Builder(allocator)
	volumes := array.NewFloat64Builder(allocator)
	counts := array.NewInt64Builder(allocator)
	sources := array.NewStringBuilder(allocator)
	defer symbols.Release()
	defer starts.Release()
	defer ends.Release()
	defer opens.Release()
	defer highs.Release()
	defer lows.Release()
	defer closes.Release()
	defer volumes.Release()
	defer counts.Release()
	defer sources.Release()

	for _, candle := range candles {
		symbols.Append(candle.Symbol)
		starts.Append(toArrowTimestamp(candle.WindowStart))
		ends.Append(toArrowTimestamp(candle.WindowEnd))
		opens.Append(candle.Open)
		highs.Append(candle.High)
		lows.Append(candle.Low)
		closes.Append(candle.Close)
		volumes.Append(candle.Volume)
		counts.Append(int64(candle.Count))
		sources.Append(candle.Source)
	}

	columns := []arrow.Array{
		symbols.NewArray(),
		starts.NewArray(),
		ends.NewArray(),
		opens.NewArray(),
		highs.NewArray(),
		lows.NewArray(),
		closes.NewArray(),
		volumes.NewArray(),
		counts.NewArray(),
		sources.NewArray(),
	}
	for _, column := range columns {
		defer column.Release()
	}

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "symbol", Type: arrow.BinaryTypes.String},
		{Name: "window_start", Type: &arrow.TimestampType{Unit: arrow.Microsecond, TimeZone: "UTC"}},
		{Name: "window_end", Type: &arrow.TimestampType{Unit: arrow.Microsecond, TimeZone: "UTC"}},
		{Name: "open", Type: arrow.PrimitiveTypes.Float64},
		{Name: "high", Type: arrow.PrimitiveTypes.Float64},
		{Name: "low", Type: arrow.PrimitiveTypes.Float64},
		{Name: "close", Type: arrow.PrimitiveTypes.Float64},
		{Name: "volume", Type: arrow.PrimitiveTypes.Float64},
		{Name: "count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "source", Type: arrow.BinaryTypes.String},
	}, nil)
	record := array.NewRecord(schema, columns, int64(len(candles)))
	defer record.Release()

	streamWriter := ipc.NewWriter(writer, ipc.WithSchema(schema))
	defer streamWriter.Close()
	return streamWriter.Write(record)
}

func toArrowTimestamp(value time.Time) arrow.Timestamp {
	return arrow.Timestamp(value.UTC().UnixMicro())
}

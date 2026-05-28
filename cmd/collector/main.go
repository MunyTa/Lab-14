package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/MunyTa/Lab-14/internal/coordination"
	"github.com/MunyTa/Lab-14/internal/market"
	"github.com/MunyTa/Lab-14/internal/rustvalidator"
	"github.com/MunyTa/Lab-14/internal/transport"
)

func main() {
	var (
		symbolsFlag = flag.String("symbols", "BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT", "comma-separated crypto symbols")
		workers     = flag.Int("workers", 1, "total collector workers")
		workerID    = flag.Int("worker-id", 0, "current worker index")
		window      = flag.Duration("window", time.Minute, "tumbling aggregation window")
		maxRecords  = flag.Int("max-records", 0, "emit a candle after this many ticks inside a window; 0 disables the limit")
		ticks       = flag.Int("ticks", 180, "number of simulated ticks per assigned symbol")
		interval    = flag.Duration("interval", 0, "delay between simulated ticks")
		outPath     = flag.String("out", "out/candles.jsonl", "JSONL output path")
		batchPath   = flag.String("batch", "out/candles_recordbatch.json", "columnar batch output path")
		etcdURL     = flag.String("etcd", "", "optional etcd endpoint, for example http://localhost:2379")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	symbols := splitSymbols(*symbolsFlag)
	plan, err := market.NewShardPlan(symbols, *workers)
	must(err)

	assigned := plan.SymbolsForWorker(*workerID)
	if len(assigned) == 0 {
		fmt.Printf("worker %d has no symbols assigned\n", *workerID)
		return
	}

	must(coordination.RegisterAssignment(ctx, *etcdURL, coordination.Assignment{
		WorkerID:    fmt.Sprintf("worker-%d", *workerID),
		WorkerIndex: *workerID,
		WorkerCount: plan.WorkerCount(),
		Symbols:     assigned,
	}))

	candles := collect(ctx, assigned, *window, *maxRecords, *ticks, *interval)
	sortCandles(candles)
	must(transport.WriteCandlesJSONL(*outPath, candles))
	must(transport.WriteBatchJSON(*batchPath, candles))

	fmt.Printf("worker=%d symbols=%s candles=%d jsonl=%s batch=%s\n", *workerID, strings.Join(assigned, ","), len(candles), *outPath, *batchPath)
}

func collect(ctx context.Context, symbols []string, window time.Duration, maxRecords, ticks int, interval time.Duration) []market.Candle {
	aggregator := market.NewWindowAggregator(window, maxRecords)
	simulator := market.NewSimulator(symbols, time.Now().UTC().Truncate(time.Second))
	candles := make([]market.Candle, 0, len(symbols)*ticks)

	for tickIndex := 0; tickIndex < ticks; tickIndex++ {
		select {
		case <-ctx.Done():
			return append(candles, aggregator.Flush()...)
		default:
		}

		for _, tick := range simulator.Next() {
			if err := rustvalidator.ValidateTick(tick); err != nil {
				continue
			}
			candles = append(candles, aggregator.Add(tick)...)
		}
		if interval > 0 {
			timer := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return append(candles, aggregator.Flush()...)
			case <-timer.C:
			}
		}
	}
	return append(candles, aggregator.Flush()...)
}

func splitSymbols(raw string) []string {
	parts := strings.Split(raw, ",")
	symbols := make([]string, 0, len(parts))
	for _, part := range parts {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol != "" {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

func sortCandles(candles []market.Candle) {
	sort.Slice(candles, func(i, j int) bool {
		if candles[i].WindowStart.Equal(candles[j].WindowStart) {
			return candles[i].Symbol < candles[j].Symbol
		}
		return candles[i].WindowStart.Before(candles[j].WindowStart)
	})
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

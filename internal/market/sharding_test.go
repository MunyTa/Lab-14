package market

import "testing"

func TestShardPlanAssignsSymbolsOnce(t *testing.T) {
	plan, err := NewShardPlan([]string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"}, 3)
	if err != nil {
		t.Fatalf("unexpected shard plan error: %v", err)
	}

	seen := map[string]int{}
	for worker := 0; worker < 3; worker++ {
		for _, symbol := range plan.SymbolsForWorker(worker) {
			seen[symbol]++
		}
	}

	for _, symbol := range []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"} {
		if seen[symbol] != 1 {
			t.Fatalf("symbol %s assigned %d times, want exactly once", symbol, seen[symbol])
		}
	}
}

func TestShardPlanRejectsInvalidWorkers(t *testing.T) {
	if _, err := NewShardPlan([]string{"BTCUSDT"}, 0); err == nil {
		t.Fatal("expected invalid worker count to return an error")
	}
}


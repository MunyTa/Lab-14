package market

import (
	"errors"
	"strings"
)

type ShardPlan struct {
	assignments [][]string
}

func NewShardPlan(symbols []string, workers int) (*ShardPlan, error) {
	if workers <= 0 {
		return nil, errors.New("worker count must be positive")
	}

	assignments := make([][]string, workers)
	for index, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		worker := index % workers
		assignments[worker] = append(assignments[worker], normalized)
	}

	return &ShardPlan{assignments: assignments}, nil
}

func (p *ShardPlan) SymbolsForWorker(worker int) []string {
	if p == nil || worker < 0 || worker >= len(p.assignments) {
		return nil
	}
	assigned := make([]string, len(p.assignments[worker]))
	copy(assigned, p.assignments[worker])
	return assigned
}

func (p *ShardPlan) WorkerCount() int {
	if p == nil {
		return 0
	}
	return len(p.assignments)
}

# Performance notes

The collector is deterministic by default, so Go and Python implementations can be compared with the same symbols and tick count.

Recommended run:

```powershell
$env:GOTELEMETRY='off'
go run ./cmd/collector -ticks 10000 -symbols BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT -out out/go_candles.jsonl
python scripts/python_collector.py --ticks 10000 --out out/python_ticks.jsonl
```

Metrics to compare:

- elapsed time printed by each collector;
- number of emitted rows;
- generated JSONL size;
- CPU and memory from the container runtime or OS task monitor.

The Go collector performs tumbling-window aggregation before writing data, so it emits fewer rows than the Python tick producer. That matches the elevated-complexity requirement to reduce data volume before the Python analysis stage.


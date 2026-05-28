# Performance report

Measurements were collected locally on 2026-05-28 for variant 8 with 4 symbols and 50,000 simulated ticks per symbol.

## Go vs Python collection

| Collector | Rows written | Elapsed, ms | Rows/s | Peak RSS, MB | CPU seconds | Output bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Go collector | 3336 | 109.99 | 30330.77 | 50.82 | 0.05 | 723028 |
| Python asyncio collector | 200000 | 1725.09 | 115936.31 | 22.08 | 1.66 | 18468520 |

The Go collector writes aggregated OHLCV candles, while the Python collector writes raw ticks. This demonstrates the expected reduced downstream volume after tumbling-window aggregation in Go.

## JSONL vs Apache Arrow IPC

| Transport | Size, bytes | Read time, ms |
| --- | ---: | ---: |
| JSONL | 723028 | 23.48 |
| Arrow IPC stream | 401568 | 13.99 |

Arrow IPC write time: 25.99 ms.  
pyarrow version: 24.0.0.

Charts:

- `reports/performance_go_python.svg`
- `reports/arrow_vs_json_size.svg`

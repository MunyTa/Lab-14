#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
import time
from pathlib import Path
from typing import Any

import pyarrow as pa


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_INPUT = ROOT / "out" / "candles.jsonl"
DEFAULT_OUTPUT = ROOT / "out" / "candles.arrow"


def main() -> None:
    input_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DEFAULT_INPUT
    output_path = Path(sys.argv[2]) if len(sys.argv) > 2 else DEFAULT_OUTPUT
    started = time.perf_counter()
    table = table_from_jsonl(input_path)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    write_arrow_stream(table, output_path)
    elapsed_ms = (time.perf_counter() - started) * 1000
    print(f"arrow_rows={table.num_rows} arrow_bytes={output_path.stat().st_size} elapsed_ms={elapsed_ms:.2f} out={output_path}")


def table_from_jsonl(path: Path) -> pa.Table:
    rows = read_rows(path)
    return pa.table(
        {
            "symbol": [row["symbol"] for row in rows],
            "window_start": [row["window_start"] for row in rows],
            "window_end": [row["window_end"] for row in rows],
            "open": [float(row["open"]) for row in rows],
            "high": [float(row["high"]) for row in rows],
            "low": [float(row["low"]) for row in rows],
            "close": [float(row["close"]) for row in rows],
            "volume": [float(row["volume"]) for row in rows],
            "count": [int(row["count"]) for row in rows],
            "source": [row.get("source", "unknown") for row in rows],
        }
    )


def write_arrow_stream(table: pa.Table, path: Path) -> None:
    with path.open("wb") as sink:
        with pa.ipc.new_stream(sink, table.schema) as writer:
            writer.write_table(table)


def read_arrow_stream(path: Path) -> pa.Table:
    with path.open("rb") as source:
        return pa.ipc.open_stream(source).read_all()


def read_rows(path: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if line:
                rows.append(json.loads(line))
    return rows


if __name__ == "__main__":
    main()

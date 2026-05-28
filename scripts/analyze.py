#!/usr/bin/env python3
from __future__ import annotations

import json
import math
import sys
import time
from collections import defaultdict
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_INPUT = ROOT / "out" / "candles.jsonl"
REPORT_DIR = ROOT / "reports" / "generated"


def main() -> None:
    input_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DEFAULT_INPUT
    candles = read_candles(input_path)
    cleaned = clean_candles(candles)
    summary = aggregate(cleaned)

    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    summary_path = REPORT_DIR / "summary.json"
    summary_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    write_bar_svg(REPORT_DIR / "close_by_symbol.svg", "Average close", [row["symbol"] for row in summary], [row["avg_close"] for row in summary])
    write_bar_svg(REPORT_DIR / "volume_by_symbol.svg", "Total volume", [row["symbol"] for row in summary], [row["total_volume"] for row in summary])

    print("first rows:")
    for row in cleaned[:5]:
        print(json.dumps(row, ensure_ascii=False))
    print(f"rows={len(cleaned)} symbols={len(summary)} summary={summary_path}")

    optional_polars_duckdb(input_path)


def read_candles(path: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if line:
                rows.append(json.loads(line))
    return rows


def clean_candles(rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    seen: set[tuple[Any, Any, Any]] = set()
    cleaned: list[dict[str, Any]] = []
    for row in rows:
        key = (row.get("symbol"), row.get("window_start"), row.get("window_end"))
        if key in seen:
            continue
        seen.add(key)
        if not row.get("symbol") or number(row.get("close")) <= 0 or number(row.get("volume")) < 0:
            continue
        cleaned.append(row)
    return cleaned


def aggregate(rows: list[dict[str, Any]]) -> list[dict[str, Any]]:
    grouped: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for row in rows:
        grouped[str(row["symbol"])].append(row)

    summary: list[dict[str, Any]] = []
    for symbol, values in sorted(grouped.items()):
        closes = [number(row["close"]) for row in values]
        highs = [number(row["high"]) for row in values]
        lows = [number(row["low"]) for row in values]
        volumes = [number(row["volume"]) for row in values]
        summary.append(
            {
                "symbol": symbol,
                "candles": len(values),
                "avg_close": round(sum(closes) / len(closes), 4),
                "min_low": round(min(lows), 4),
                "max_high": round(max(highs), 4),
                "total_volume": round(sum(volumes), 4),
            }
        )
    return summary


def optional_polars_duckdb(input_path: Path) -> None:
    started = time.perf_counter()
    try:
        import duckdb  # type: ignore
        import polars as pl  # type: ignore
    except ImportError:
        print("polars/duckdb are not installed; stdlib analysis was used")
        return

    frame = pl.read_ndjson(input_path)
    frame = frame.unique(subset=["symbol", "window_start", "window_end"]).drop_nulls()
    parquet_path = REPORT_DIR / "candles.parquet"
    frame.write_parquet(parquet_path)

    query = """
        SELECT symbol,
               COUNT(*) AS candles,
               AVG(close) AS avg_close,
               MIN(low) AS min_low,
               MAX(high) AS max_high,
               SUM(volume) AS total_volume
        FROM read_parquet(?)
        WHERE close > 0
        GROUP BY symbol
        ORDER BY total_volume DESC
    """
    rows = duckdb.sql(query, params=[str(parquet_path)]).fetchall()
    elapsed_ms = (time.perf_counter() - started) * 1000
    print(f"duckdb_rows={len(rows)} parquet={parquet_path} elapsed_ms={elapsed_ms:.2f}")


def write_bar_svg(path: Path, title: str, labels: list[str], values: list[float]) -> None:
    width = 860
    height = 420
    margin = 60
    chart_height = height - 2 * margin
    chart_width = width - 2 * margin
    maximum = max(values) if values else 1
    bar_width = chart_width / max(len(values), 1)
    bars: list[str] = []

    for index, (label, value) in enumerate(zip(labels, values)):
        bar_height = 0 if maximum == 0 else (value / maximum) * chart_height
        x = margin + index * bar_width + 12
        y = height - margin - bar_height
        bars.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{bar_width - 24:.1f}" height="{bar_height:.1f}" fill="#2f80ed"/>')
        bars.append(f'<text x="{x + (bar_width - 24) / 2:.1f}" y="{height - 28}" font-size="13" text-anchor="middle">{label}</text>')
        bars.append(f'<text x="{x + (bar_width - 24) / 2:.1f}" y="{max(y - 8, 20):.1f}" font-size="12" text-anchor="middle">{value:.2f}</text>')

    svg = f"""<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">
  <rect width="100%" height="100%" fill="#ffffff"/>
  <text x="{width / 2}" y="32" font-size="22" font-family="Arial" text-anchor="middle">{title}</text>
  <line x1="{margin}" y1="{height - margin}" x2="{width - margin}" y2="{height - margin}" stroke="#222"/>
  <line x1="{margin}" y1="{margin}" x2="{margin}" y2="{height - margin}" stroke="#222"/>
  <g font-family="Arial">{''.join(bars)}</g>
</svg>
"""
    path.write_text(svg, encoding="utf-8")


def number(value: Any) -> float:
    try:
        result = float(value)
    except (TypeError, ValueError):
        return math.nan
    if math.isnan(result) or math.isinf(result):
        return 0
    return result


if __name__ == "__main__":
    main()


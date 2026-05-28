#!/usr/bin/env python3
from __future__ import annotations

import argparse
import asyncio
import json
from collections import defaultdict, deque
from datetime import datetime, timedelta, timezone
from typing import Any


async def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", type=int, default=4222)
    parser.add_argument("--subject", default="lab14.crypto.candles")
    parser.add_argument("--window-seconds", type=int, default=300)
    args = parser.parse_args()

    reader, writer = await asyncio.open_connection(args.host, args.port)
    greeting = await reader.readline()
    if not greeting.startswith(b"INFO "):
        raise RuntimeError(f"unexpected NATS greeting: {greeting!r}")

    writer.write(b'CONNECT {"verbose":false,"pedantic":false}\r\n')
    writer.write(f"SUB {args.subject} 1\r\nPING\r\n".encode())
    await writer.drain()

    windows: dict[str, deque[dict[str, Any]]] = defaultdict(deque)
    while True:
        line = (await reader.readline()).decode().strip()
        if not line or line == "PONG":
            continue
        if not line.startswith("MSG "):
            continue

        parts = line.split()
        size = int(parts[-1])
        payload = await reader.readexactly(size)
        await reader.readexactly(2)
        candle = json.loads(payload)
        update_window(windows[candle["symbol"]], candle, args.window_seconds)
        print(json.dumps(rolling_summary(windows), ensure_ascii=False))


def update_window(rows: deque[dict[str, Any]], candle: dict[str, Any], window_seconds: int) -> None:
    rows.append(candle)
    current = parse_time(candle["window_end"])
    edge = current - timedelta(seconds=window_seconds)
    while rows and parse_time(rows[0]["window_end"]) < edge:
        rows.popleft()


def rolling_summary(windows: dict[str, deque[dict[str, Any]]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for symbol, rows in windows.items():
        if not rows:
            continue
        result[symbol] = {
            "candles": len(rows),
            "avg_close": round(sum(float(row["close"]) for row in rows) / len(rows), 4),
            "total_volume": round(sum(float(row["volume"]) for row in rows), 4),
        }
    return result


def parse_time(value: str) -> datetime:
    return datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(timezone.utc)


if __name__ == "__main__":
    asyncio.run(main())

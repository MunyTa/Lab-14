#!/usr/bin/env python3
from __future__ import annotations

import argparse
import asyncio
import json
import math
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path


@dataclass
class Tick:
    symbol: str
    price: float
    quantity: float
    timestamp: float


async def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--symbols", default="BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT")
    parser.add_argument("--ticks", type=int, default=180)
    parser.add_argument("--out", default="out/python_ticks.jsonl")
    args = parser.parse_args()

    symbols = [item.strip().upper() for item in args.symbols.split(",") if item.strip()]
    started = time.perf_counter()
    output = Path(args.out)
    output.parent.mkdir(parents=True, exist_ok=True)

    with output.open("w", encoding="utf-8") as handle:
        for step in range(args.ticks):
            for tick in await produce(symbols, step):
                handle.write(json.dumps(tick.__dict__) + "\n")

    elapsed_ms = (time.perf_counter() - started) * 1000
    print(f"python_collector rows={len(symbols) * args.ticks} elapsed_ms={elapsed_ms:.2f} out={output}")


async def produce(symbols: list[str], step: int) -> list[Tick]:
    await asyncio.sleep(0)
    timestamp = datetime.now(timezone.utc).timestamp() + step
    ticks: list[Tick] = []
    for index, symbol in enumerate(symbols):
        price = 1000 + (index + 1) * 250 + math.sin(step + index) * 12.5 + (step % 17) * 0.8
        quantity = 0.25 + ((step + index) % 9) * 0.15
        ticks.append(Tick(symbol=symbol, price=round(price, 4), quantity=round(quantity, 4), timestamp=timestamp))
    return ticks


if __name__ == "__main__":
    asyncio.run(main())


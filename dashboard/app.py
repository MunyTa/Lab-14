from __future__ import annotations

import asyncio
import json
from pathlib import Path
from typing import Any

from fastapi import FastAPI, WebSocket
from fastapi.responses import HTMLResponse


ROOT = Path(__file__).resolve().parents[1]
DATA_PATH = ROOT / "out" / "candles.jsonl"

app = FastAPI(title="Lab 14 Crypto Dashboard")


@app.get("/")
async def index() -> HTMLResponse:
    return HTMLResponse(
        """
<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>Crypto dashboard</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 32px; color: #172033; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border-bottom: 1px solid #d8dee9; padding: 8px 10px; text-align: right; }
    th:first-child, td:first-child { text-align: left; }
  </style>
</head>
<body>
  <h1>Crypto candles</h1>
  <table>
    <thead><tr><th>Symbol</th><th>Window</th><th>Open</th><th>High</th><th>Low</th><th>Close</th><th>Volume</th></tr></thead>
    <tbody id="rows"></tbody>
  </table>
  <script>
    const rows = document.getElementById("rows");
    const socket = new WebSocket(`ws://${location.host}/ws`);
    socket.onmessage = event => {
      const candles = JSON.parse(event.data);
      rows.innerHTML = candles.slice(-40).reverse().map(c => `
        <tr>
          <td>${c.symbol}</td><td>${new Date(c.window_start).toLocaleTimeString()}</td>
          <td>${c.open.toFixed(4)}</td><td>${c.high.toFixed(4)}</td>
          <td>${c.low.toFixed(4)}</td><td>${c.close.toFixed(4)}</td><td>${c.volume.toFixed(4)}</td>
        </tr>`).join("");
    };
  </script>
</body>
</html>
        """
    )


@app.get("/api/candles")
async def candles(limit: int = 100) -> list[dict[str, Any]]:
    return load_candles(limit)


@app.websocket("/ws")
async def websocket(websocket: WebSocket) -> None:
    await websocket.accept()
    while True:
        await websocket.send_json(load_candles(100))
        await asyncio.sleep(2)


def load_candles(limit: int) -> list[dict[str, Any]]:
    if not DATA_PATH.exists():
        return []
    rows: list[dict[str, Any]] = []
    with DATA_PATH.open("r", encoding="utf-8") as handle:
        for line in handle:
            if line.strip():
                rows.append(json.loads(line))
    return rows[-limit:]


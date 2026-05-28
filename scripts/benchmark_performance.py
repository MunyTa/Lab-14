#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import subprocess
import sys
import time
import ctypes
from dataclasses import dataclass
from pathlib import Path

import pyarrow as pa

from arrow_ipc import read_arrow_stream, table_from_jsonl, write_arrow_stream


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "out"
REPORTS = ROOT / "reports"


@dataclass
class RunMetric:
    name: str
    command: str
    elapsed_ms: float
    cpu_seconds: float | None
    peak_rss_mb: float | None
    rows: int
    bytes_written: int

    @property
    def rows_per_second(self) -> float:
        return self.rows / max(self.elapsed_ms / 1000, 0.000001)


def main() -> None:
    OUT.mkdir(exist_ok=True)
    REPORTS.mkdir(exist_ok=True)

    env = os.environ.copy()
    env["GOTELEMETRY"] = "off"
    env["GOCACHE"] = str(ROOT / ".gocache")
    env["GOMODCACHE"] = str(ROOT / ".gomodcache")

    collector_exe = ROOT / "bin" / ("collector.exe" if os.name == "nt" else "collector")
    collector_exe.parent.mkdir(exist_ok=True)
    subprocess.run(["go", "build", "-o", str(collector_exe), "./cmd/collector"], cwd=ROOT, env=env, check=True)

    go_json = OUT / "bench_go_candles.jsonl"
    python_json = OUT / "bench_python_ticks.jsonl"
    arrow_path = OUT / "bench_go_candles.arrow"
    ticks = "50000"
    symbols = "BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT"

    go_metric = run_measured(
        "Go collector",
        [str(collector_exe), "-ticks", ticks, "-symbols", symbols, "-out", str(go_json), "-batch", str(OUT / "bench_batch.json")],
        env,
        go_json,
    )
    python_metric = run_measured(
        "Python asyncio collector",
        [sys.executable, str(ROOT / "scripts" / "python_collector.py"), "--ticks", ticks, "--symbols", symbols, "--out", str(python_json)],
        env,
        python_json,
    )

    arrow_started = time.perf_counter()
    table = table_from_jsonl(go_json)
    write_arrow_stream(table, arrow_path)
    arrow_write_ms = (time.perf_counter() - arrow_started) * 1000

    json_read_ms = measure(lambda: table_from_jsonl(go_json))
    arrow_read_ms = measure(lambda: read_arrow_stream(arrow_path))

    write_report(go_metric, python_metric, go_json, arrow_path, arrow_write_ms, json_read_ms, arrow_read_ms, pa.__version__)
    write_bar_svg(REPORTS / "performance_go_python.svg", "Rows per second", [go_metric.name, python_metric.name], [go_metric.rows_per_second, python_metric.rows_per_second])
    write_bar_svg(REPORTS / "arrow_vs_json_size.svg", "Transport size, bytes", ["JSONL", "Arrow IPC"], [go_json.stat().st_size, arrow_path.stat().st_size])
    print(REPORTS / "performance.md")


def run_measured(name: str, command: list[str], env: dict[str, str], output_path: Path) -> RunMetric:
    started = time.perf_counter()
    process = subprocess.Popen(command, cwd=ROOT, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    peak_rss = 0
    cpu_seconds: float | None = None

    while process.poll() is None:
        sample = sample_process(process.pid)
        if sample is not None:
            rss, cpu = sample
            peak_rss = max(peak_rss, rss)
            cpu_seconds = cpu
        time.sleep(0.05)

    stdout, stderr = process.communicate()
    elapsed_ms = (time.perf_counter() - started) * 1000
    if process.returncode != 0:
        raise RuntimeError(f"{name} failed\nSTDOUT:\n{stdout}\nSTDERR:\n{stderr}")

    rows = count_lines(output_path)
    return RunMetric(
        name=name,
        command=" ".join(command),
        elapsed_ms=elapsed_ms,
        cpu_seconds=cpu_seconds,
        peak_rss_mb=peak_rss / (1024 * 1024) if peak_rss else None,
        rows=rows,
        bytes_written=output_path.stat().st_size,
    )


def sample_process(pid: int) -> tuple[int, float | None] | None:
    if os.name != "nt":
        return None
    return sample_windows_process(pid)


def sample_windows_process(pid: int) -> tuple[int, float | None] | None:
    process_query_information = 0x0400
    process_vm_read = 0x0010
    handle = ctypes.windll.kernel32.OpenProcess(process_query_information | process_vm_read, False, pid)
    if not handle:
        return None
    try:
        counters = ProcessMemoryCounters()
        counters.cb = ctypes.sizeof(ProcessMemoryCounters)
        if not ctypes.windll.psapi.GetProcessMemoryInfo(handle, ctypes.byref(counters), counters.cb):
            return None

        creation = FileTime()
        exit_time = FileTime()
        kernel = FileTime()
        user = FileTime()
        cpu_seconds = None
        if ctypes.windll.kernel32.GetProcessTimes(handle, ctypes.byref(creation), ctypes.byref(exit_time), ctypes.byref(kernel), ctypes.byref(user)):
            cpu_seconds = (filetime_to_int(kernel) + filetime_to_int(user)) / 10_000_000
        return int(counters.WorkingSetSize), cpu_seconds
    finally:
        ctypes.windll.kernel32.CloseHandle(handle)


class FileTime(ctypes.Structure):
    _fields_ = [("dwLowDateTime", ctypes.c_uint32), ("dwHighDateTime", ctypes.c_uint32)]


class ProcessMemoryCounters(ctypes.Structure):
    _fields_ = [
        ("cb", ctypes.c_uint32),
        ("PageFaultCount", ctypes.c_uint32),
        ("PeakWorkingSetSize", ctypes.c_size_t),
        ("WorkingSetSize", ctypes.c_size_t),
        ("QuotaPeakPagedPoolUsage", ctypes.c_size_t),
        ("QuotaPagedPoolUsage", ctypes.c_size_t),
        ("QuotaPeakNonPagedPoolUsage", ctypes.c_size_t),
        ("QuotaNonPagedPoolUsage", ctypes.c_size_t),
        ("PagefileUsage", ctypes.c_size_t),
        ("PeakPagefileUsage", ctypes.c_size_t),
    ]


def filetime_to_int(value: FileTime) -> int:
    return (int(value.dwHighDateTime) << 32) + int(value.dwLowDateTime)


def count_lines(path: Path) -> int:
    with path.open("r", encoding="utf-8") as handle:
        return sum(1 for line in handle if line.strip())


def measure(callback) -> float:
    started = time.perf_counter()
    callback()
    return (time.perf_counter() - started) * 1000


def write_report(
    go_metric: RunMetric,
    python_metric: RunMetric,
    json_path: Path,
    arrow_path: Path,
    arrow_write_ms: float,
    json_read_ms: float,
    arrow_read_ms: float,
    pyarrow_version: str,
) -> None:
    metrics = {
        "pyarrow_version": pyarrow_version,
        "go": go_metric.__dict__,
        "python": python_metric.__dict__,
        "json_bytes": json_path.stat().st_size,
        "arrow_bytes": arrow_path.stat().st_size,
        "arrow_write_ms": arrow_write_ms,
        "json_read_ms": json_read_ms,
        "arrow_read_ms": arrow_read_ms,
    }
    (REPORTS / "performance.json").write_text(json.dumps(metrics, indent=2), encoding="utf-8")

    content = f"""# Performance report

Measurements were collected locally on 2026-05-28 for variant 8 with 4 symbols and 50,000 simulated ticks per symbol.

## Go vs Python collection

| Collector | Rows written | Elapsed, ms | Rows/s | Peak RSS, MB | CPU seconds | Output bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Go collector | {go_metric.rows} | {go_metric.elapsed_ms:.2f} | {go_metric.rows_per_second:.2f} | {fmt(go_metric.peak_rss_mb)} | {fmt(go_metric.cpu_seconds)} | {go_metric.bytes_written} |
| Python asyncio collector | {python_metric.rows} | {python_metric.elapsed_ms:.2f} | {python_metric.rows_per_second:.2f} | {fmt(python_metric.peak_rss_mb)} | {fmt(python_metric.cpu_seconds)} | {python_metric.bytes_written} |

The Go collector writes aggregated OHLCV candles, while the Python collector writes raw ticks. This demonstrates the expected reduced downstream volume after tumbling-window aggregation in Go.

## JSONL vs Apache Arrow IPC

| Transport | Size, bytes | Read time, ms |
| --- | ---: | ---: |
| JSONL | {json_path.stat().st_size} | {json_read_ms:.2f} |
| Arrow IPC stream | {arrow_path.stat().st_size} | {arrow_read_ms:.2f} |

Arrow IPC write time: {arrow_write_ms:.2f} ms.  
pyarrow version: {pyarrow_version}.

Charts:

- `reports/performance_go_python.svg`
- `reports/arrow_vs_json_size.svg`
"""
    (REPORTS / "performance.md").write_text(content, encoding="utf-8")


def fmt(value: float | None) -> str:
    return "n/a" if value is None else f"{value:.2f}"


def write_bar_svg(path: Path, title: str, labels: list[str], values: list[float]) -> None:
    width = 760
    height = 360
    margin = 60
    maximum = max(values) if values else 1
    bar_width = (width - 2 * margin) / max(len(values), 1)
    bars = []
    for index, (label, value) in enumerate(zip(labels, values)):
        bar_height = 0 if maximum == 0 else (value / maximum) * (height - 2 * margin)
        x = margin + index * bar_width + 20
        y = height - margin - bar_height
        bars.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{bar_width - 40:.1f}" height="{bar_height:.1f}" fill="#1b7f5a"/>')
        bars.append(f'<text x="{x + (bar_width - 40) / 2:.1f}" y="{height - 25}" text-anchor="middle" font-size="13">{label}</text>')
        bars.append(f'<text x="{x + (bar_width - 40) / 2:.1f}" y="{max(y - 8, 24):.1f}" text-anchor="middle" font-size="12">{value:.2f}</text>')
    path.write_text(
        f"""<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">
<rect width="100%" height="100%" fill="white"/>
<text x="{width / 2}" y="32" text-anchor="middle" font-size="22" font-family="Arial">{title}</text>
<line x1="{margin}" y1="{height - margin}" x2="{width - margin}" y2="{height - margin}" stroke="#222"/>
<line x1="{margin}" y1="{margin}" x2="{margin}" y2="{height - margin}" stroke="#222"/>
<g font-family="Arial">{''.join(bars)}</g>
</svg>
""",
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()

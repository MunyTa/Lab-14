# Лабораторная работа №14

ФИО: Кузьмищев Родион Ильич  
Группа: 221331  
Вариант: 8  
Тема варианта: мониторинг криптовалют Binance/Bybit WebSocket  
Уровень выполнения: повышенная сложность

## Что реализовано

Проект адаптирует вариант 8 под задания повышенной сложности. Вместо сырых JSON-тиков сборщик на Go распределяет список криптовалютных пар между worker-экземплярами, агрегирует поток в tumbling-window OHLCV-свечи и сохраняет результат в JSONL и columnar RecordBatch-представление для Python/Arrow-слоя.

Ключевые части:

- распределение символов между несколькими Go-сборщиками;
- optional регистрация назначений worker-ов в etcd через `/v3/kv/put`;
- оконная агрегация OHLCV в Go до передачи в Python;
- бинарный Apache Arrow IPC stream через `pyarrow` и optional Go Arrow HTTP server;
- Rust-библиотека валидации с C ABI и Go cgo-интеграцией через build tag `rustvalidate`;
- потоковая передача агрегированных свечей через NATS;
- Python consumer со скользящим окном поверх NATS-потока;
- Python-анализ с очисткой, агрегацией, SVG-графиками и optional Polars/DuckDB;
- FastAPI + WebSocket dashboard для просмотра последних свечей;
- Docker Compose с etcd и двумя collector-экземплярами;
- Kubernetes Deployment и HPA для автоскалирования.

## Архитектура

```text
crypto stream / simulator
        |
        v
Go collector workers -> sharding -> optional etcd assignment registry
        |
        v
tumbling OHLCV aggregation
        |
        +--> out/candles.jsonl
        +--> out/candles.arrow (Apache Arrow IPC)
        +--> NATS subject lab14.crypto.candles
        |
        v
Python analysis / sliding-window consumer -> reports/*.json + *.svg
        |
        v
FastAPI dashboard
```

По умолчанию используется детерминированный эмулятор потока. Это позволяет проверять тесты и анализ без доступа к Binance/Bybit и без API-ключей.

## Запуск

Тесты:

```powershell
$env:GOTELEMETRY='off'
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
$env:GOMODCACHE=(Join-Path (Get-Location) '.gomodcache')
go test ./...
```

Сбор данных:

```powershell
go run ./cmd/collector -symbols BTCUSDT,ETHUSDT,SOLUSDT,BNBUSDT -workers 2 -worker-id 0 -ticks 180 -window 1m
```

Анализ:

```powershell
python scripts/analyze.py out/candles.jsonl
```

Apache Arrow IPC:

```powershell
python scripts/arrow_ipc.py out/candles.jsonl out/candles.arrow
```

Optional Go Arrow HTTP server:

```powershell
go mod tidy
go run -tags arrow ./cmd/arrow-server -input out/candles.jsonl -addr :8081
```

NATS-поток:

```powershell
docker compose -f deploy/docker-compose.yml up nats
go run ./cmd/collector -nats-url nats://localhost:4222 -ticks 180
python scripts/nats_sliding_consumer.py --window-seconds 300
```

Rust-валидация через cgo:

```powershell
cd rust_validator
cargo build --release
cd ..
go run -tags rustvalidate ./cmd/collector
```

Performance benchmark:

```powershell
python scripts/benchmark_performance.py
```

Полный Python-стек для Polars, DuckDB, pyarrow и dashboard:

```powershell
python -m pip install -r requirements.txt
uvicorn dashboard.app:app --reload
```

Docker Compose:

```powershell
docker compose -f deploy/docker-compose.yml up --build
```

Kubernetes:

```powershell
kubectl apply -f deploy/k8s/collector-deployment.yaml
kubectl apply -f deploy/k8s/collector-hpa.yaml
```

## Проверка повышенной сложности

- `internal/market/window.go` выполняет оконную агрегацию на стороне Go.
- `internal/market/sharding.go` делит пары между несколькими сборщиками.
- `internal/coordination/etcd.go` регистрирует назначения worker-ов в etcd.
- `scripts/arrow_ipc.py` пишет настоящий бинарный Apache Arrow IPC stream.
- `cmd/arrow-server/main.go` отдает Arrow IPC stream по HTTP при сборке с `-tags arrow`.
- `rust_validator/src/lib.rs` содержит Rust-валидацию, `internal/rustvalidator/validator_cgo.go` подключает ее через cgo.
- `internal/streaming/nats.go` публикует свечи в NATS, `scripts/nats_sliding_consumer.py` считает скользящее окно.
- `reports/performance.md` содержит фактические замеры Go vs Python и JSONL vs Arrow.
- `scripts/analyze.py` чистит данные, агрегирует результаты и строит 2 SVG-графика.
- `dashboard/app.py` показывает свежие свечи через WebSocket.
- `deploy/k8s/collector-hpa.yaml` содержит HPA для автоскалирования.

## Git-процесс

Работа выполнена через поэтапный TDD-поток:

1. добавлены красные тесты для агрегации, шардирования, валидации и batch-представления;
2. реализовано Go-ядро до зеленых тестов;
3. добавлены CLI, анализ, dashboard, deployment и документация;
4. по итогам review добавлены Arrow IPC, Rust/cgo, NATS streaming, фактический performance report и `PROMPT_LOG.md`;
5. после каждого этапа выполнен отдельный commit и push.

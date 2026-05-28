# PROMPT_LOG

## Контекст

Лабораторная работа №14 выполнялась с помощью AI-инструмента Codex в локальном рабочем каталоге `C:\Users\Kuzmi\OneDrive\Desktop\Лабораторная 14`.

Студент: Кузьмищев Родион Ильич  
Группа: 221331  
Вариант: 8  
Предметная область: мониторинг криптовалют  
Требуемый уровень: повышенная сложность

## Ход работы с AI

1. AI прочитал методичку из PDF и выделил список заданий повышенной сложности.
2. Для варианта 8 была адаптирована предметная область: криптовалютный поток Binance/Bybit заменен детерминированным эмулятором, чтобы проект стабильно проверялся без API-ключей.
3. Работа началась с TDD: сначала были добавлены красные тесты на агрегацию OHLCV, шардирование, валидацию и batch-представление.
4. После прохождения тестов AI реализовал Go-ядро: модели тиков/свечей, tumbling-window агрегацию, шардирование worker-ов и запись результатов.
5. После первой проверки review bot указал критические дефекты: отсутствие настоящего Arrow IPC, Rust-интеграции, NATS/Kafka-потока, фактического performance report и этого файла.
6. Во второй итерации AI добавил бинарный Apache Arrow IPC через `pyarrow`, optional Go Arrow HTTP server, Rust crate с C ABI и Go cgo wrapper, NATS publisher, Python sliding-window consumer и фактический отчет с замерами.
7. AI запускал тесты, сборщик, Arrow export, Python-компиляцию и benchmark; результаты внесены в `reports/performance.md`.
8. Git-история велась через поэтапные коммиты и push в `https://github.com/MunyTa/Lab-14`.

## Основные промпты

- Выполнить лабораторную работу №14 для варианта 8 на повышенную сложность.
- Создать README с ФИО, группой, вариантом и номером лабораторной.
- Использовать атомарные коммиты и push после этапов: красные тесты, реализация, запуск проверок.
- Исправить замечания review bot: реализовать Apache Arrow, Rust-валидацию, performance report, NATS/Kafka streaming и PROMPT_LOG.

## Проверки

- `go test -count=1 ./...`
- `go run ./cmd/collector ...`
- `python scripts/analyze.py out/candles.jsonl`
- `python scripts/arrow_ipc.py out/candles.jsonl out/candles.arrow`
- `python scripts/benchmark_performance.py`
- `python -m py_compile scripts/analyze.py scripts/python_collector.py scripts/arrow_ipc.py scripts/benchmark_performance.py scripts/nats_sliding_consumer.py dashboard/app.py`

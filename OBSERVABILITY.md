# Observability

В проекте настроен мониторинг сервисов `bot` и `scrapper` с использованием Prometheus и Grafana.

Prometheus собирает метрики с `/metrics` endpoint-ов приложений, а Grafana отображает dashboard с бизнес-метриками и техническими метриками потребления памяти.

## Подключенные инструменты

В мониторинг входят следующие компоненты:

- `Prometheus` — сбор и хранение метрик;
- `Grafana` — визуализация метрик;
- `/metrics` endpoint в `scrapper`;
- отдельный `/metrics` endpoint в `bot`;
- Grafana dashboard `Link Tracker Business Metrics`;
- Prometheus datasource для Grafana.

## Метрики Scrapper

### `links_on_track_total`

Тип: `GaugeVec`.

Labels:

- `app="scrapper"` — приложение;
- `tracked_source` — источник отслеживаемой ссылки.

Описание: показывает количество активных ссылок в БД, поставленных на мониторинг, с разделением по источнику.

Метрика используется для панели количества активных отслеживаемых ссылок.

### `request_duration_ms_total`

Тип: `HistogramVec`.

Labels:

- `app="scrapper"` — приложение;
- `scope` — область операции;
- `scope_type` — тип операции внутри области.

Описание: хранит распределение времени выполнения операций Scrapper в миллисекундах.

Buckets:

```text
10, 25, 50, 100, 250, 500, 1000, 2500, 5000
```

Метрика используется для расчета p50, p95 и p99 времени выполнения операций через `histogram_quantile`.

### `api_requests_total`

Тип: `CounterVec`.

Labels:

- `app="scrapper"` — приложение;
- `source` — источник API-запроса.

Описание: считает количество API-запросов, обработанных Scrapper.

### `api_errors_total`

Тип: `CounterVec`.

Labels:

- `app="scrapper"` — приложение;
- `source` — источник ошибки.

Описание: считает количество ошибок API-запросов в Scrapper.

## Метрики Bot

### `sent_notification_total`

Тип: `Counter`.

Labels:

- `app="bot"` — приложение.

Описание: считает количество отправленных пользователям уведомлений.

### `command_duration_ms_total`

Тип: `HistogramVec`.

Labels:

- `app="bot"` — приложение;
- `scope` — область операции;
- `scope_type` — тип операции.

Описание: хранит распределение времени обработки операций Bot в миллисекундах.

Buckets:

```text
10, 25, 50, 100, 250, 500, 1000, 2500, 5000
```

Метрика используется для расчета p50, p95 и p99 времени обработки команд.

### `command_requests_total`

Тип: `CounterVec`.

Labels:

- `app="bot"` — приложение;
- `command` — название обработанной команды.

Описание: считает количество команд или пользовательских сообщений, обработанных Bot.

### `errors_counter_total`

Тип: `CounterVec`.

Labels:

- `app="bot"` — приложение;
- `scope` — область ошибки;
- `scope_type` — тип операции, в которой произошла ошибка.

Описание: считает количество ошибок при обработке операций в Bot.

## Стандартные Go/process метрики

Кроме кастомных метрик, Prometheus client экспортирует стандартные Go/process метрики.

Они используются для отображения потребления памяти приложениями:

- `process_resident_memory_bytes` — фактическое потребление RAM процессом;
- `go_memstats_alloc_bytes` — выделенная Go runtime память;
- `go_memstats_stack_inuse_bytes` — память, используемая стеками goroutine.

## PromQL запросы из dashboard

Ниже перечислены PromQL-запросы, на основе которых построены панели dashboard.

### 1. Количество пользовательских сообщений в секунду

```promql
sum(rate(command_requests_total{app="bot"}[1m]))
```

Описание: показывает среднее количество пользовательских сообщений или команд, обработанных Bot за секунду за последнюю минуту.

Метрика: `command_requests_total`.

Тип визуализации: `Time series`.

### 2. Количество активных ссылок в БД по источнику

```promql
links_on_track_total{app="scrapper"}
```

Описание: показывает количество активных ссылок, которые находятся на мониторинге, с разделением по label `tracked_source`.

Метрика: `links_on_track_total`.

Тип визуализации: `Time series`.

### 3. p50 времени операции Scrapper

```promql
histogram_quantile(
  0.50,
  sum(rate(request_duration_ms_total_bucket{app="scrapper", scope="database"}[5m])) by (le, scope_type)
)
```

Описание: показывает 50-й перцентиль времени выполнения операций Scrapper за последние 5 минут с разделением по `scope_type`.

Метрика: `request_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

### 4. p95 времени операции Scrapper

```promql
histogram_quantile(
  0.95,
  sum(rate(request_duration_ms_total_bucket{app="scrapper", scope="database"}[5m])) by (le, scope_type)
)
```

Описание: показывает 95-й перцентиль времени выполнения операций Scrapper за последние 5 минут с разделением по `scope_type`.

Метрика: `request_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

### 5. p99 времени операции Scrapper

```promql
histogram_quantile(
  0.99,
  sum(rate(request_duration_ms_total_bucket{app="scrapper", scope="database"}[5m])) by (le, scope_type)
)
```

Описание: показывает 99-й перцентиль времени выполнения операций Scrapper за последние 5 минут с разделением по `scope_type`.

Метрика: `request_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

### 6. p50 времени обработки команды в Bot

```promql
histogram_quantile(
  0.50,
  sum(rate(command_duration_ms_total_bucket{app="bot"}[5m])) by (le, scope, scope_type)
)
```

Описание: показывает 50-й перцентиль времени обработки операций Bot за последние 5 минут.

Метрика: `command_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

### 7. p95 времени обработки команды в Bot

```promql
histogram_quantile(
  0.50,
  sum(rate(command_duration_ms_total_bucket{app="bot"}[5m])) by (le, scope, scope_type)
)
```

Описание: запрос используется в dashboard для второй линии панели Bot command duration percentiles. В текущей версии dashboard он совпадает с p50.

Метрика: `command_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

> Примечание: если требуется именно p95, значение `0.50` в запросе нужно заменить на `0.95`.

### 8. p99 времени обработки команды в Bot

```promql
histogram_quantile(
  0.99,
  sum(rate(command_duration_ms_total_bucket{app="bot"}[5m])) by (le, scope, scope_type)
)
```

Описание: показывает 99-й перцентиль времени обработки операций Bot за последние 5 минут.

Метрика: `command_duration_ms_total_bucket`.

Тип визуализации: `Time series`.

### 9. Общее количество запросов к боту в Telegram

```promql
sum(command_requests_total{app="bot"})
```

Описание: показывает суммарное количество пользовательских команд или сообщений, обработанных Bot.

Метрика: `command_requests_total`.

Тип визуализации: `Stat`.

### 10. Количество отправленных нотификаций в секунду

```promql
sum(rate(sent_notification_total{app="bot"}[1m]))
```

Описание: показывает среднее количество отправленных уведомлений в секунду за последнюю минуту.

Метрика: `sent_notification_total`.

Тип визуализации: `Stat`.

### 11. Потребление RAM приложением

```promql
process_resident_memory_bytes{job="$app"} / 1024 / 1024
```

Описание: показывает фактическое потребление RAM процессом в мегабайтах. Значение `$app` выбирается через переменную dashboard.

Метрика: `process_resident_memory_bytes`.

Тип визуализации: `Stat`.

### 12. Go allocated memory

```promql
go_memstats_alloc_bytes{job="$app"} / 1024 / 1024
```

Описание: показывает количество памяти, выделенной Go runtime, в мегабайтах.

Метрика: `go_memstats_alloc_bytes`.

Тип визуализации: `Stat`.

### 13. Go stack memory

```promql
go_memstats_stack_inuse_bytes{job="$app"} / 1024 / 1024
```

Описание: показывает количество памяти, используемой стеками goroutine, в мегабайтах.

Метрика: `go_memstats_stack_inuse_bytes`.

Тип визуализации: `Stat`.

## Dashboard filter

В Grafana dashboard добавлена переменная `$app`.

Значения переменной:

- `bot`;
- `scrapper`.

Она используется для фильтрации панелей по типу приложения, например в запросах потребления памяти:

```promql
process_resident_memory_bytes{job="$app"} / 1024 / 1024
```

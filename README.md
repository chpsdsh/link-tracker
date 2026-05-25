# LinkTracker

**LinkTracker** – Telegram-бот, который отслеживает изменения на веб-страницах и оперативно информирует пользователя о них.

## Запуск
```yaml
  docker compose up -d
```
**Для запуска необходимо:** 
1) **Создать файл "bot.env" с полями:**
```yaml
   APP_TELEGRAM_TOKEN=token
   SCRAPPER_SERVER_ADDRESS=http://scrapper:8081
   WITH_TELEGRAM_API=true
   UPDATES_HANDLE_TYPE=(kafka/http/fallback)
   KAFKA_USER=user1
   KAFKA_PASSWORD=user1-secret
   KAFKA_RAW_TOPIC=raw-notification-topic
   KAFKA_PROCESSED_TOPIC=processed-notification-topic
   KAFKA_DLQ_TOPIC=notification-dlq
   KAFKA_CONSUMER_GROUP=processed-notification-consumers
   KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
   POSTGRES_HOST=postgres
   POSTGRES_PORT=5432
   POSTGRES_USER=user
   POSTGRES_PASSWORD=password
   POSTGRES_DB=mydb
   HTTP_CLIENT_TIMEOUT=10s
   RETRY_MAX_ATTEMPTS=3
   RETRY_DELAY=500ms
   RETRYABLE_STATUSES=500,502,503,504
   CIRCUIT_BREAKER_INTERVAL=10s
   CIRCUIT_BREAKER_TIMEOUT=5s
   CIRCUIT_BREAKER_MAX_REQUESTS=3
   CIRCUIT_BREAKER_FAILURE_RATIO=0.6
   RATE_LIMIT_RPS=5
   RATE_LIMIT_BURST=10
```
 
2) **Создать файл "scrapper.env" с полями:**
```yaml
   GITHUB_API_KEY=key
   STACKOVERFLOW_API_KEY=key
   BOT_SERVER_ADDRESS=http://bot:8080
   SCRAPPER_TIME_INTERVAL=10
   LINKS_BATCH_SIZE=20
   SCHEDULER_NUM_WORKERS=5
   ASSET_TYPE=BUILDER
   POSTGRES_HOST=postgres
   POSTGRES_PORT=5432
   POSTGRES_USER=user
   POSTGRES_PASSWORD=password
   POSTGRES_DB=mydb
   UPDATES_HANDLE_TYPE=(kafka/http/fallback)
   KAFKA_USER=user1
   KAFKA_PASSWORD=user1-secret
   KAFKA_TOPIC=raw-notification-topic
   KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
   VALKEY_PASSWORD=valkey
   VALKEY_ADDRESSES=valkey-node-0:6379
   VALKEY_TTL_MINUTES=5
   HTTP_CLIENT_TIMEOUT=10s
   RETRY_MAX_ATTEMPTS=3
   RETRY_DELAY=500ms
   RETRYABLE_STATUSES=500,502,503,504
   CIRCUIT_BREAKER_INTERVAL=10s
   CIRCUIT_BREAKER_TIMEOUT=5s
   CIRCUIT_BREAKER_MAX_REQUESTS=3
   CIRCUIT_BREAKER_FAILURE_RATIO=0.6
   RATE_LIMIT_RPS=5
   RATE_LIMIT_BURST=10
```
3) **Создать файл "agent.env" с полями:**
```yaml
   KAFKA_USER=user1
   KAFKA_PASSWORD=user1-secret
   KAFKA_RAW_TOPIC=raw-notification-topic
   KAFKA_PROCESSED_TOPIC=processed-notification-topic
   KAFKA_DLQ_TOPIC=notification-dlq
   KAFKA_CONSUMER_GROUP=raw-notification-consumers
   KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
   POSTGRES_HOST=postgres
   POSTGRES_PORT=5432
   POSTGRES_USER=user
   POSTGRES_PASSWORD=password
   POSTGRES_DB=mydb
   AI_STOP_WORDS=spam,ads,promo
   AI_EXCLUDED_AUTHORS=bot-user,spam-author
   AI_MIN_LENGTH=20
   AI_SUMMARIZATION_THRESHOLD=500
   AI_HIGH_PRIORITY_KEY_WORDS=critical,urgent,breaking
   GROUP_WINDOW_MS=30000
   AI_LOW_PRIORITY_KEY_WORDS=minor,typo,docs,chore
   YANDEX_API_KEY=key(API ключ от Yandex AI Studio)
   YANDEX_FOLDER_ID=key(id папки в Yandex AI Studio)
   YANDEX_MODEL=yandexgpt-5-lite
   YANDEX_BASE_URL=https://ai.api.cloud.yandex.net/v1
```
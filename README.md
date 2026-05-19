# LinkTracker

**LinkTracker** – Telegram-бот, который отслеживает изменения на веб-страницах и оперативно информирует пользователя о них.

## Запуск
```yaml
  docker compose up -d
```
**Для запуска необходимо:** 
1) **Создать файл "bot.env" с полями:**
```yaml
   APP_TELEGRAM_TOKEN=key
   SCRAPPER_SERVER_ADDRESS=http://scrapper:8081
   WITH_TELEGRAM_API=true
   UPDATES_RECEIVER_TYPE=kafka
   KAFKA_USER=user1
   KAFKA_PASSWORD=user1-secret
   KAFKA_TOPIC=notification-topic
   KAFKA_DLQ_TOPIC=notification-dlq
   KAFKA_CONSUMER_GROUP=notification-consumers
   KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
   POSTGRES_HOST=postgres
   POSTGRES_PORT=5432
   POSTGRES_USER=user
   POSTGRES_PASSWORD=password
   POSTGRES_DB=mydb
   HTTP_CLIENT_TIMEOUT_SEC=10
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
   UPDATES_SEND_TYPE=kafka
   KAFKA_USER=user1
   KAFKA_PASSWORD=user1-secret
   KAFKA_TOPIC=notification-topic
   KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
   VALKEY_PASSWORD=valkey
   VALKEY_ADDRESSES=valkey-node-0:6379
   VALKEY_TTL_MINUTES=5
   HTTP_CLIENT_TIMEOUT_SEC=10
```
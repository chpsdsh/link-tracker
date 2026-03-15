# LinkTracker

**LinkTracker** – Telegram-бот, который отслеживает изменения на веб-страницах и оперативно информирует пользователя о них.

## Запуск
**Для запуска необходимо:** 
1) **Создать файл "bot.env" с полями:**
 - APP_TELEGRAM_TOKEN=token
 - SCRAPPER_SERVER_ADDRESS=http://localhost:8081
 - WITH_TELEGRAM_API=true (false нужно ТОЛЬКО для интеграциионных тестов)
2) **Создать файл "scrapper.env" с полями:**
 - GITHUB_API_KEY=token
 - STACKOVERFLOW_API_KEY=token
 - BOT_SERVER_ADDRESS=http://localhost:8080
3) **Для интеграционных тестов соберите образы**
 - docker build -f scrapper.dockerfile -t scrapper-image .
 - docker build -f bot.dockerfile -t bot-image .
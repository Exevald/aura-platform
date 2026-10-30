# Auth service

Сервис управляет учётными данными и аутентификацией.

## Модель и транзакции

- Домен содержит агрегат, бизнес-правила, интерфейс репозитория и доменные события.
- AuthService: регистрация, вход, logout и смена пароля.
- App-слой открывает UnitOfWork, создаёт доменный сервис с транзакционным репозиторием и диспетчером событий и выполняет операцию.
- MySQL-репозиторий хранит данные в таблице credential. Изменение модели и запись события в outbox выполняются в одной транзакции.

Публичный HTTP API предоставляет `/register`, `/login`, `/logout` и `/change-password` под префиксом `/api/public/v1/auth`. Auth синхронно вызывает внутренние gRPC API user и session. Пароли длиной 8–128 Unicode-символов хешируются Argon2id.

Session ID хранится в HttpOnly/Secure/SameSite=Strict cookie. Без `rememberMe` cookie непостоянная, а серверная сессия живёт 8 часов; с `rememberMe` cookie и серверная сессия живут 30 дней. `AUTH_COOKIE_DOMAIN` задаёт общий доверенный домен для продуктов на поддоменах; локально значение оставляют пустым и используют один hostname, например `localhost`.

## Запуск и конфигурация

Бинарник поддерживает команду service (по умолчанию). Она применяет миграции, запускает HTTP, gRPC и outbox relay.

Основные переменные окружения используют префикс AUTH_: MYSQL_*, HTTP_ADDRESS, GRPC_ADDRESS, AMQP_*, RELAY_*, DEPENDENCIES_* и COOKIE_*.

Из корня репозитория:

    mise run //modules/auth:generate-migration "Describe migration"
    mise run //modules/auth:build
    mise run //modules/auth:test
    mise run //modules/auth:integration-tests

Интеграционный тест поднимает user, auth и session с отдельными MySQL и RabbitMQ и проверяет регистрацию, повторную регистрацию, login, logout, remember-me, смену пароля и публикацию доменных событий.

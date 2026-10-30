# Session service

Сервис управляет пользовательскими сессиями.

## Модель и транзакции

- Домен содержит агрегат, бизнес-правила, интерфейс репозитория и доменные события.
- SessionService: создание, чтение, валидация и отзыв одной или всех сессий пользователя.
- Обычная сессия живёт 8 часов; remember-me сессия — 30 дней. Оба срока проверяются сервером.
- App-слой открывает UnitOfWork, создаёт доменный сервис с транзакционным репозиторием и диспетчером событий и выполняет операцию.
- MySQL-репозиторий хранит данные в таблице session. Изменение модели и запись события в outbox выполняются в одной транзакции.

Внутренний gRPC API предоставляет CreateSession, ValidateSession, RevokeSession и RevokeUserSessions. HTTP-сервер предоставляет /healthz. События SessionCreated и SessionRevoked публикуются relay-процессом в общий exchange domain-event.

## Запуск и конфигурация

Бинарник поддерживает команду service (по умолчанию). Она применяет миграции, запускает HTTP, gRPC и outbox relay.

Основные переменные окружения используют префикс SESSION_: MYSQL_*, HTTP_ADDRESS, GRPC_ADDRESS, AMQP_*, RELAY_*, DEFAULT_LIFETIME и PERSISTENT_LIFETIME.

Из корня репозитория:

    mise run //modules/session:generate-migration "Describe migration"
    mise run //modules/session:build
    mise run //modules/session:test
    mise run //modules/session:integration-tests

Интеграционный тест проверяет создание, валидацию и отзыв обычной и remember-me сессий через gRPC.

# Access service

Сервис управляет доступом пользователей к продуктовым модулям.

## Модель и транзакции

- Домен содержит агрегат, бизнес-правила, интерфейс репозитория и доменные события.
- ModuleAccessService: создание, чтение, изменение и отзыв доступа, а также проверка пары user/module.
- App-слой открывает UnitOfWork, создаёт доменный сервис с транзакционным репозиторием и диспетчером событий и выполняет операцию.
- MySQL-репозиторий хранит данные в таблице module_access. Изменение модели и запись события в outbox выполняются в одной транзакции.

Публичных бизнес-контрактов HTTP и gRPC пока нет. HTTP-сервер предоставляет /healthz, gRPC-сервер поддерживает reflection. События публикуются relay-процессом в общий exchange domain-event; сервис не создаёт собственную очередь.

## Запуск и конфигурация

Бинарник поддерживает команду service (по умолчанию). Она применяет миграции, запускает HTTP, gRPC и outbox relay.

Основные переменные окружения используют префикс ACCESS_: MYSQL_*, HTTP_ADDRESS, GRPC_ADDRESS, AMQP_URL, AMQP_CONNECT_TIMEOUT и RELAY_*.

Из корня репозитория:

    mise run //modules/access:generate-migration "Describe migration"
    mise run //modules/access:build
    mise run //modules/access:test
    mise run //modules/access:integration-tests

Интеграционный тест запускает сервис с MySQL и RabbitMQ и проверяет health endpoint и применение бизнес-миграции.

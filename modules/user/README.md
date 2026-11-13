# User service

Сервис управляет пользовательскими профилями.

## Модель и транзакции

- Доменная модель User содержит display name, имя, фамилию и email. Credential login хранится только в auth и не связан с display name.
- Доменный UserService создаёт, изменяет и удаляет пользователей через репозиторий и отправляет события UserCreated, UserUpdated и UserRemoved.
- App-слой открывает UnitOfWork, внутри транзакции создаёт доменный сервис с репозиторием и outbox-диспетчером и выполняет операцию.
- MySQL-репозиторий хранит профили в таблице user; данные и событие фиксируются атомарно.

## API

Внутренний gRPC API предоставляет CreateUser, GetUser, UpdateUser и DeleteUser, а также XA prepare/commit/rollback для регистрации. Публичного API создания профиля нет: регистрацией и финальным XA-решением управляет registerer.

HTTP-сервер предоставляет /healthz. Relay публикует доменные события в exchange domain-event; сервис не создаёт собственную очередь.

## Запуск

Бинарник поддерживает команду service (по умолчанию). Основные переменные окружения используют префикс USER_: MYSQL_*, HTTP_ADDRESS, GRPC_ADDRESS, AMQP_URL, AMQP_CONNECT_TIMEOUT и RELAY_*.

    mise run //modules/user:generate-migration "Describe migration"
    mise run //modules/user:build
    mise run //modules/user:test
    mise run //modules/user:integration-tests

Интеграционный тест проверяет полный CRUD профиля через gRPC и его сохранение в MySQL.

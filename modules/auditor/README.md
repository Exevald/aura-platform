# Auditor service

Сервис хранит неизменяемые записи аудита действий платформы.

## Модель и транзакции

- Доменная модель AuditRecord фиксирует субъект, действие, ресурс, метаданные и время события.
- AuditService создаёт и читает записи через транзакционный репозиторий.
- App-слой открывает UnitOfWork, создаёт доменный сервис и выполняет операцию.
- MySQL-репозиторий хранит данные в таблице audit_record.

Сервис пока не публикует и не принимает события, поэтому ему не нужны RabbitMQ, outbox relay и отдельный consumer. Публичных бизнес-контрактов HTTP и gRPC пока нет: HTTP-сервер предоставляет /healthz, gRPC-сервер поддерживает reflection.

Бинарник поддерживает команду service (по умолчанию). Основные переменные окружения используют префикс AUDITOR_: MYSQL_*, HTTP_ADDRESS, GRPC_ADDRESS и APP_*.

    mise run //modules/auditor:generate-migration "Describe migration"
    mise run //modules/auditor:build
    mise run //modules/auditor:test
    mise run //modules/auditor:integration-tests

Интеграционный тест запускает сервис с MySQL и проверяет health endpoint и применение миграции audit_record.

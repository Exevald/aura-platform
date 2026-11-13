# Notification service

Хранит пользовательский inbox, read-state и настройки каналов `in_app/email`.
Внутренний gRPC API принимает идемпотентные команды создания; запись уведомления и
`NotificationCreated` фиксируются одной транзакцией через outbox.

```sh
mise run //modules/notification:test
mise run //modules/notification:integration-tests
```

Интеграционный тест проверяет gRPC-команду, чтение inbox и изменение preferences.

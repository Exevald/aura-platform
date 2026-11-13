# Auth service

Сервис владеет credentials, проверяет пароль и обслуживает login, logout и смену пароля.

## API и транзакции

Публичный HTTP API предоставляет `/login`, `/logout` и `/change-password` под
префиксом `/api/public/v1/auth`. Регистрация вынесена в `registerer`.

Для регистрации auth предоставляет внутренние gRPC-методы
`PrepareCredentialRegistration`, `CommitCredentialRegistration` и
`RollbackCredentialRegistration`. Credential и локальные outbox-записи
подготавливаются как отдельная MySQL XA-ветка; финальное решение принимает
`registerer`.

Пароли длиной 8–128 Unicode-символов хешируются Argon2id. Auth синхронно вызывает
только session. Session ID передаётся в HttpOnly/Secure/SameSite=Strict cookie.

## Запуск

Основные переменные окружения используют префикс `AUTH_`: `MYSQL_*`,
`HTTP_ADDRESS`, `GRPC_ADDRESS`, `AMQP_*`, `RELAY_*`,
`DEPENDENCIES_SESSION_ADDRESS` и `COOKIE_*`. MySQL должен использовать InnoDB и
`xa_detach_on_prepare=ON`.

```sh
mise run //modules/auth:build
mise run //modules/auth:test
mise run //modules/auth:integration-tests
```

Интеграционный тест проверяет XA prepare/commit credential, duplicate login,
login, logout, remember-me, смену пароля и публикацию auth/session событий.

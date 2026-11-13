# Registerer service

Registerer владеет сценарием регистрации и является координатором двухфазного
коммита между `auth`, `user` и `access`.

## Сценарий

Публичный endpoint `POST /api/public/v1/registerer/register` принимает login,
password, профиль и `rememberMe`. Registerer создаёт `transactionID` и `userID`,
записывает состояние `preparing`, вызывает XA prepare участников и сохраняет
необратимое решение `committing` либо `aborting` в собственной MySQL.

Recovery worker повторяет commit/rollback после сетевых ошибок и перезапуска.
После завершения всех веток registerer атомарно переводит операцию в
`committed` и записывает `UserRegistered` в outbox. Создание session выполняется
после commit auth-ветки и не входит в 2PC: при ошибке регистрация остаётся
успешной, а ответ содержит `authenticated=false`.

## Конфигурация

Основные переменные используют префикс `REGISTERER_`:

- `MYSQL_*`, `AMQP_*`, `RELAY_*`;
- `RECOVERY_INTERVAL`, `RECOVERY_BATCH_SIZE`;
- `DEPENDENCIES_AUTH_ADDRESS`, `USER_ADDRESS`, `ACCESS_ADDRESS`,
  `SESSION_ADDRESS`, `EDITOR_MODULE_ID`, `TIMEOUT`;
- `COOKIE_NAME`, `COOKIE_DOMAIN`, `COOKIE_SECURE`.

```sh
mise run //modules/registerer:build
mise run //modules/registerer:test
mise run //e2e:integration-tests
```

Полный e2e проверяет commit, abort, потерю prepare-ответа, повторную доставку
решения, перезапуск координатора, duplicate login и независимость регистрации от
session.

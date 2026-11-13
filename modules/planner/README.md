# Planner service

Workspace-aware проекты и задачи. Задача содержит статус, приоритет, исполнителя,
срок и optimistic version; опционально хранит Calendar ID и связанное событие.
Изменения проекта и задачи атомарно публикуются через outbox.

```sh
mise run //modules/planner:test
mise run //modules/planner:integration-tests
```

Интеграционный тест проверяет project/task CRUD и stale-version conflict.

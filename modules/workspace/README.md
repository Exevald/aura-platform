# Workspace service

Владеет личными и командными рабочими пространствами и ролями `owner/editor/viewer`.
HTTP API предоставляет список, создание, чтение и управление участниками. Внутренний
gRPC API идемпотентно создаёт personal workspace и проверяет membership для gateway.

Данные хранятся в `workspace` и `workspace_member`; изменения публикуются через
outbox как `WorkspaceCreated`, `WorkspaceMembershipChanged` и `WorkspaceDeleted`.

```sh
mise run //modules/workspace:test
mise run //modules/workspace:integration-tests
```

Интеграционный тест проверяет создание personal workspace, membership check,
создание team workspace и чтение списка через HTTP.

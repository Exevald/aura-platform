# UserGroup

Сервис управляет группами и их участниками внутри выбранного workspace. Группы принадлежат владельцу и `workspace_id`; публичный API требует `X-Auth-User` и `X-Workspace-ID`. Пользователи разрешаются через сервис `user` по email.

Membership-события публикуются через transactional outbox. `access` использует внутренний gRPC API для получения состава группы и идемпотентно проецирует назначения модулей на её участников.

Проверка: `mise run //modules/usergroup:integration-tests`.

# Realtime service

Держит WebSocket-соединения в памяти и доставляет адресные события пользователю
или workspace. Gateway передаёт проверенные `X-Auth-User` и `X-Workspace-ID`;
внутренний gRPC `Publish` используется Notification и продуктовыми сервисами.
Медленные клиенты не блокируют hub, соединения поддерживаются ping/pong.

```sh
mise run //modules/realtime:test
mise run //modules/realtime:integration-tests
```

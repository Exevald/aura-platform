# Calendar service

Workspace-aware календари и события. HTTP API поддерживает CRUD календарей и
событий, IANA timezone, RRULE `DAILY/WEEKLY/MONTHLY`, участников, несколько
напоминаний и optimistic version. Внутренний gRPC API позволяет Planner
идемпотентно создавать/обновлять и удалять связанное событие.

MySQL хранит `calendar`, `calendar_event`, attendees и reminders. Изменения
публикуются через outbox как `CalendarChanged` и `EventChanged`.

```sh
mise run //modules/calendar:test
mise run //modules/calendar:integration-tests
```

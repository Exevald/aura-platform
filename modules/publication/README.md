# Publication

Отдельный сервис публикации сохранённых версий документов Editor. Публикация создаёт неизменяемую публичную копию
Markdown и вложений, короткую ссылку Shortlink и уведомляет участников рабочего пространства. Снятие публикации и
удаление исходного документа закрывают доступ и запускают очистку.

## Контракт и сага

`api/workflow/publication.proto` описывает Temporal workflow `Publish`/`Unpublish`, activities, task queue, таймауты и
retry. `protoc-gen-go-temporal` генерирует типизированный клиент и регистрацию worker. Бизнес-порядок и компенсация
находятся в `internal/publication/app/flow/workflow.go`.

```text
Publish:   ExportSnapshot → CreateShortLink → Activate → NotifyMembers
                           при ошибке до Activate: DeleteShortLink → DeleteSnapshot
Unpublish: Close → DeleteShortLink → DeleteSnapshot
```

После `Activate` страница уже доступна. Ошибка уведомлений не отменяет публикацию: статус доставки становится `failed`.
Идемпотентность запуска задаёт заголовок `Idempotency-Key`; Temporal workflow ID и Shortlink ID также стабильны для
каждой публикации. Копии вложений хранятся через Filestorage, текст и состояние — в MySQL. На старте сервис повторно запускает
незавершённые процессы; удаление документа поступает через событие `editor.DocumentChanged`.

## API

- `POST /api/publication/v1/releases` с `Idempotency-Key` и JSON `{"documentID":"<uuid>","expectedVersion":1}` —
  запустить публикацию (202).
- `GET /api/publication/v1/releases` — список публикаций владельца.
- `GET /api/publication/v1/releases/{id}` — статус.
- `DELETE /api/publication/v1/releases/{id}` — снять публикацию (204).
- `GET /p/{id}` и `GET /p/{id}/assets/{fileID}` — анонимный просмотр активной копии.

Управляющие маршруты проходят через gateway с проверкой сессии, рабочего пространства и доступа к модулю. Публичные
маршруты доступны без сессии. Ссылка Shortlink ведёт на `/p/{id}` и принадлежит техническому владельцу модуля; повторная
публикация создаёт новую ссылку.

## Сборка и запуск

Из корня репозитория:

```sh
mise install
mise run //modules/publication:generate
mise run //modules/publication:build
mise run local-up
```

`mise` фиксирует версии Buf и `protoc-gen-go-temporal`. Зависимые `.proto` находятся в `temporal/proto`; generated Go,
как и в остальных модулях Aura, не хранится в Git. Команды `editorAdapter`, `worker` и `message-handler` запускаются
отдельными контейнерами. Temporal UI доступен на `http://localhost:8080`.

Основные настройки: `PUBLICATION_TEMPORAL_ADDRESS`, `PUBLICATION_TEMPORAL_NAMESPACE`, `PUBLICATION_DEPENDENCIES_FILESTORAGE_ADDRESS`,
`PUBLICATION_DEPENDENCIES_*`, `PUBLICATION_MYSQL_*`. Локальные значения заданы в `config/config.go` и `compose.yaml`.

## Проверки сценария

`mise run //modules/publication:integration-tests` поднимает изолированные MySQL, RabbitMQ и Temporal, сервис и
message handler. Worker работает в тестовом процессе; Editor, Filestorage, Shortlink, Workspace и Notification заменены управляемыми
gRPC-сервисами. Так проверяются настоящие выполнение workflow, хранилища и HTTP API с воспроизводимыми ошибками внешних
модулей.

| Интеграционный тест                  | Ожидаемый результат                                                                                                                                                                                                                          |
|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Публикация документа с вложением     | Страница и байты вложения доступны анонимно; короткая ссылка создана; уведомления доставлены владельцу и участнику. Повтор с тем же ключом возвращает прежнюю публикацию без новых побочных эффектов, а повтор с другим документом даёт 409. |
| Изменение версии во время публикации | Workflow завершается ошибкой, страница остаётся закрытой, ссылка и экспортированные объекты удаляются.                                                                                                                                       |
| Ошибка Shortlink                     | Workflow компенсирует экспорт вложения и не открывает страницу.                                                                                                                                                                              |
| Ошибка Notification после активации  | Страница остаётся открытой, статус доставки становится `failed`.                                                                                                                                                                             |
| Снятие публикации                    | Страница и вложение отвечают 404; ссылка и копия файла в Filestorage удаляются.                                                                                                                                                              |

После `mise run local-up` команда `mise run //e2e:publication-e2e` проходит через gateway и реальные сервисы. Тест
регистрирует владельца и участника, выдаёт доступ к Editor и Publication, создаёт документ и публикацию, проверяет
разграничение доступа к управлению, анонимную страницу, переход по Shortlink и уведомления. Затем он изменяет исходный
документ и проверяет неизменность опубликованного снимка, снимает публикацию, публикует новую версию и удаляет документ:
последняя публикация должна закрыться автоматически. Адрес gateway задаётся через `AURA_E2E_BASE_URL` (по умолчанию
`http://127.0.0.1` в задаче mise).

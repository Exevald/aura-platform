# Editor

Редактор хранит Markdown-документы, оптимистическую версию, ресурсные роли owner/editor/viewer, групповые публикации и связи документов с вложениями. Содержимое документа остаётся в MySQL; бинарные вложения передаются streaming gRPC в `filestorage` и физически хранятся в MinIO.

Публичный API расположен под `/api/editor/v1`. Для вложений доступны multipart upload, список, download и delete под `/api/editor/v1/documents/{documentID}/attachments`. Gateway проверяет сессию, `X-Workspace-ID` и доступ к модулю; Editor дополнительно проверяет роль конкретного документа.

Проверка: `mise run //modules/editor:integration-tests`.

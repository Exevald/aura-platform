# Filestorage service

Метаданные файлов и SHA-256 хранятся в MySQL, содержимое — в S3-совместимом
MinIO. Multipart upload и streaming download проходят через HTTP/gateway;
внутренний gRPC API поддерживает streaming upload/download для Editor.

```sh
mise run //modules/filestorage:test
mise run //modules/filestorage:integration-tests
```

Интеграционный тест поднимает MinIO и проверяет upload, checksum, download и delete.

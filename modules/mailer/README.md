# Mailer service

Внутренний gRPC API принимает идемпотентные SMTP-задания. MySQL хранит recipient,
subject/body, status, attempts и последнюю ошибку; успешная или неуспешная доставка
публикуется через outbox. Локальный SMTP — Mailpit.

```sh
mise run //modules/mailer:test
mise run //modules/mailer:integration-tests
```

Интеграционный тест отправляет письмо в Mailpit и проверяет подавление повторной
команды с тем же idempotency key.

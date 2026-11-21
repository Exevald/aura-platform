# aura-platform

Платформа независимых Go-сервисов для совместной работы: документы, календари,
задачи, короткие ссылки, публикации, рабочие пространства, уведомления и аудит.

## Проверка и локальное окружение

Нужны mise, Python 3, Docker и Compose. Поддерживаются как плагин
`docker compose`, так и отдельная команда `docker-compose`.

```sh
mise trust
mise install
mise run local-up
```

При первом запуске `local-up` попросит пароль администратора, чтобы добавить локальные
адреса Aura в `/etc/hosts`. Затем задача находит все `modules/*/compose.yaml`,
генерирует контракты и собирает Linux-бинарники через модульные задачи mise.
Фронтенд собирается внутри Docker-образа с `npm ci`; отдельная локальная сборка не нужна.
После этого задача запускает общую инфраструктуру,
создаёт базы и пользователей из модульных `docker/mysql/init.sql` и поднимает
сервисы. SQL подготовки запускается при каждом старте, чтобы новые модули работали
и с уже существующим volume MySQL.

Открывайте `http://aura-platform.home/`. Продукты доступны на
`http://aura-platform.documents/`, `http://aura-platform.calendar/`,
`http://aura-platform.planner/`, `http://aura-platform.shortlinks/`,
`http://aura-platform.publications/` и
`http://aura-platform.notifications/`. Единый вход работает через
`http://aura-platform.auth/`: переход между продуктами создаёт отдельную
сессию для каждого домена по краткоживущему одноразовому коду.
Внешний HTTP-порт можно изменить через `LOCAL_WEB_PORT`; итоговые адреса
`local-up` печатает после запуска.

Корневой `compose.yaml` содержит MySQL, RabbitMQ, MinIO, Mailpit и Temporal (UI на `http://localhost:8080`). Каждый модуль описывает
свои сервисы, очередь, настройки подключения и подготовку базы у себя.

```sh
mise run local-config
mise run local-status
mise run local-logs
mise run local-down
```

`local-config` показывает объединённую конфигурацию без запуска контейнеров.
`local-down` сохраняет volumes с данными. Имя Compose-проекта задаётся через
`LOCAL_ENV_PROJECT` (по умолчанию `distributed-programming-local`).

Для первого созданного модуля по умолчанию доступны gRPC на 8081, HTTP на 8082 и health consumer
на 8083. Остальные модули получают следующие свободные порты по порядку имён;
`local-up` печатает назначенные порты. Назначения можно задать явно через
`<MODULE>_GRPC_PORT`, `<MODULE>_HTTP_PORT`, `<MODULE>_MESSAGE_HANDLER_PORT`.
Инфраструктурные порты переопределяются через `LOCAL_MYSQL_PORT`, `LOCAL_AMQP_PORT`
и `LOCAL_AMQP_MANAGEMENT_PORT`.

Продуктовые запросы идут через `gateway` и несут выбранный контекст
`X-Workspace-ID`; gateway проверяет сессию, членство в workspace и доступ к модулю.
Подробности контрактов и гарантий находятся в README соответствующих сервисов.

## Новый модуль и миграции

```sh
mise run generate-module billing
mise run //modules/billing:all
mise run //modules/billing:generate-migration "Create invoices table"
```

Исходник шаблона находится в `mise-tasks/generate-module.d` и не входит в `go.work`
или список модулей mise. Все файлы шаблона имеют суффикс `.tpl`.
Генератор записывает их в `modules/<name>` без этого суффикса, переименовывает module path, импорты,
каталог `internal/template`, protobuf-сервисы, CLI/env-префиксы и модульный Compose.
Затем генерирует Go-код из контрактов и добавляет модуль в `go.work`.
Имя должно соответствовать `[a-z][a-z0-9_]*`; существующий каталог не перезаписывается.

Генератор миграций создаёт `version<timestamp>.go` и регистрирует миграцию в
`factory.go`. Пока `Up` не реализован, миграция возвращает ошибку. Версия всегда
больше существующих, в том числе при генерации нескольких миграций в одну секунду.

Общий exchange `domain-event` описан в `lib/event`. Модули описывают свои очереди
и bindings, сохраняя общий exchange.

## Проверки

```sh
mise run all
mise run //modules/billing:integration-tests
mise run //e2e:integration-tests
mise run //e2e:publication-e2e
mise run check-generator
```

`all` выполняет генерацию, tidy, Linux-сборку, unit и интеграционные тесты, lint. Интеграционные
проверки запускаются последовательно; задача останавливается на первой ошибке.
`publication-e2e` запускается после `local-up` и проверяет публикацию через gateway и реальные сервисы; адрес задаётся `AURA_E2E_BASE_URL` (по умолчанию `http://127.0.0.1`).
Тесты находятся в `internal/integration-tests` и исключены из обычной задачи `test`.
Задача mise собирает Linux-бинарник и Docker-образ модуля. Тестовая фикстура поднимает
MySQL, RabbitMQ и при необходимости MinIO/Mailpit, сервис и consumer в общей изолированной Docker-сети.
Оба приложения запускаются из собранного образа; API доступны через случайные
порты хоста.

Общий пакет `lib/integrationtest` создаёт сеть и контейнеры, собирает логи и
освобождает ресурсы. Настройки запуска и сценарии остаются в
`internal/integration-tests` каждого сервиса. `bff` и `registerer` тоже имеют
контрактные интеграционные тесты собранных образов.

Отдельный модуль `e2e` поднимает полный контур регистрации через общий раннер
`go-sdk/pkg/integrationtests`. Он проверяет MySQL XA commit/rollback, потерю ответа
после prepare, повторную доставку commit и восстановление после перезапуска registerer, отсутствие событий у отменённой
регистрации и независимость commit от доступности session.
Для проверки используются отдельные базы и пользователи модулей в одном MySQL контейнере,
чтобы полный прогон помещался в небольшую локальную Docker VM.

Проверка генераторов работает во временной копии: создаёт `sample_echo`,
проверяет миграции, объединение Compose и полный `all`, а также отказ при
неверном имени и очистку после ошибки генерации.

## Docker

Dockerfile только копирует готовый бинарник. Сборка выполняется локально:

```sh
mise run //modules/billing:build
docker build -t distributed-programming-billing:local modules/billing
```

`build` создаёт Linux-бинарник в `dist/billing`, с отключённым CGO.
`GOARCH` позволяет выбрать архитектуру вручную; `local-up` выбирает архитектуру
Docker-сервера. Образ запускает `editorAdapter` по умолчанию; для consumer передаётся
команда `message-handler`.

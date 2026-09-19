## Архитектура aura-platform

### 1. Контекстная диаграмма

```mermaid
C4Context
    System_Ext(frontend, "Frontend", "Пользовательский интерфейс")
    System_Ext(externalClient, "Внешний клиент", "Интеграции и внешние системы")

    Enterprise_Boundary(system, "Система") {
        System(platform, "Platform", "Ядро продукта")
        System(modules, "Modules", "Независимые продуктовые модули")
    }

    Rel(frontend, platform, "Использует")
    Rel(externalClient, platform, "Использует")
    Rel(platform, modules, "Обращается к")
```

### 2. Диаграмма компонентов

```mermaid
flowchart TB
    Frontend[Frontend]
    ExternalClient[Внешний клиент]

    subgraph Platform[Platform]
        Gateway[gateway]
        BFF[bff]
        Auth[auth]
        Session[session]
        User[user]
        Workspace[workspace]
        Access[access]
        AppReg[appreg]
        Realtime[realtime]
        Notification[notification]
        Filestorage[filestorage]
        Mailer[mailer]
        Auditor[auditor]
    end

    subgraph Modules[Modules]
        Calendar[calendar]
        Editor[editor]
        Planner[planner]
    end

    Frontend --> Gateway
    ExternalClient --> Gateway
    Gateway -.->|proxy| BFF
    Gateway -.->|proxy| Auth
    Gateway -.->|proxy| Calendar
    Gateway -.->|proxy| Editor
    Gateway -.->|proxy| Planner
    Auth --> User
    Auth --> Session
    BFF --> User
    BFF --> AppReg
    BFF --> Calendar
    BFF --> Editor
    BFF --> Planner
    Access --> Workspace
    Editor --> Access
    Realtime --> Session
    Realtime --> Access
    Editor -.-> Realtime
    Calendar -.-> Realtime
    Planner -.-> Realtime
    Notification -.-> Realtime
    User --> Workspace
    Editor --> Workspace
    Editor --> Filestorage
    Calendar --> User
    Calendar -.-> Notification
    Calendar -.-> Mailer
    Planner --> Calendar
    Planner -.-> Notification
```

### 3. Описание компонентов

#### `gateway`

**Ответственность**

* единая точка входа в продукт
* принимает запросы от фронта и внешних клиентов
* направляет запрос в нужный платформенный компонент или модуль
* передаёт контекст пользователя вызываемым компонентам
* хранит в себе информацию о пользовательской сессии

**Не ответственность**

* не содержит бизнес-логику модулей
* не хранит пользовательские или продуктовые данные
* не принимает решения о бизнес-доступе к ресурсам

**Данные**

Не является владельцем бизнес-данных.

**Зависимости**

* нет

**Потребители**

* Frontend
* внешние клиенты

---

#### `user`

**Ответственность**

* логинация пользователя
* регистрация
* управление ролями и полями пользователя
* привязка пользователя к аккаунту

**Не ответственность**

* настройки аккаунта
* права доступа к продуктовым ресурсам

**Данные**

* учётные данные

**Зависимости**

* нет

**Потребители**

* `gateway`
* `calendar`
* `mailer`
* и другие продуктовые модули

---

#### `auth`

**Ответственность**

* аутентификация пользователя
* проверка кредов
* запись кредов
* связывание способа аутентификации с `user`
* запуск создания пользовательской сессии после успешного входа

**Не ответственность**

* хранить профиль пользователя
* хранить корпоративные данные
* хранить бизнес-права
* принимать решения о доступе к документам/календарям
* владеть пользовательской сессией после её создания

**Данные**

* креды пользователя
* способы аутенфикации

**Зависимости**

* `user`
* `session`

**Потребители**

* `bff`
* `gateway`

---

#### `session`

**Ответственность**

* жизненный цикл сессии

**Не ответственность**

* хранить креды
* проверять креды
* проверять права доступа

**Данные**

* данные сессии (чья, какой статус, когда закончится)

**Зависимости**

* `auth`

**Потребители**

* `auth`
* `bff`
* `realtime`

---

#### `access`

**Ответственность**

* общая платформенная авторизация
* проверка возможности пользователя выполнить действие
* хранение общих правил и разрешений

**Не ответственность**

* аутентификация
* специфические бизнес-правила отдельных модулей
* хранение ресурсов модулей

**Данные**

* разрешения
* правила доступа

**Зависимости**

* `account`

**Потребители**

* `gateway`
* `editor`
* и другие продуктовые модули

---

#### `bff`

**Ответственность**

* сбор фронтовых данных с запросов
* хранение специфичных для фронта данных

**Не ответственность**

* хранение продуктовых данных
* не проксирует каждый запрос
* не авторизует запросы

**Данные**

* нет

**Зависимости**

* нет

**Потребители**

* фронт

---

#### `realtime`

**Ответственность**

* хранение и управление долгоживущих соединений (вебсокет)
* доставка realtime-событий подключённым клиентам

**Не ответственность**

* хранение продуктовых данных
* не проксирует каждый запрос
* не авторизует запросы

**Данные**

* соединение клиента с сервером
* realtime-события

**Зависимости**

* `session`
* `access`

**Потребители**

* продуктовые сервисы, поддерживающие realtime-события

---

#### `appreg`

**Ответственность**

* реестр приложений, поддерживаемых на платформе
* данные о роутинге этих приложений

**Не ответственность**

* бизнес-логика этих приложений

**Данные**

* yaml-файлик с описанием каждого сервиса

**Зависимости**

* нет

**Потребители**

* `gateway`
* `bff`

---

#### `notification`

**Ответственность**

* отправка уведомлений пользователю

**Не ответственность**

* бизнес-логика конкретного модуля

**Данные**

* уведомления
* состояние доставки уведомления

**Зависимости**

* нет

**Потребители**

* `calendar`
* и другие продуктовые модули

---

#### `filestorage`

**Ответственность**

* единое хранилище файлов для модулей продукта

**Не ответственность**

* содержательная бизнес-логика документов
* правила доступа к документам
* данные аккаунта пользователя

**Данные**

* файлы

**Зависимости**

* нет

**Потребители**

* `editor`
* и другие продуктовые модули

---

#### `mailer`

**Ответственность**

* отправка электронной почты

**Не ответственность**

* принятие продуктовых решений о необходимости отправки письма
* хранение бизнес-сущностей модулей

**Данные**

* данные, необходимые для отправки и отслеживания писем

**Зависимости**

* `identity`

**Потребители**

* `calendar`
* и другие продуктовые модули

---

#### `auditor`

**Ответственность**

* сбор продуктовых событий
* формирование общего отчёта активностей модулей

**Не ответственность**

* бизнес-логика модулей
* авторизация
* изменение состояния бизнес-сущностей

**Данные**

* записи аудита
* сведения о произошедших действиях

**Зависимости**

* нет

**Потребители**

* `editor`
* `calendar`
* `planner`

---

#### `calendar`

**Ответственность**

* бизнес-логика календаря
* управление собственными календарными сущностями
* инициирование уведомлений и писем, связанных с календарными событиями

**Не ответственность**

* аутентификация пользователей
* реализация механизма уведомлений
* реализация отправки email

**Данные**

* календари
* события календаря
* данные, относящиеся исключительно к календарному модулю

**Зависимости**

* `identity`
* `notification`
* `mailer`
* `auditor`

**Потребители**

* `gateway`
* `planner`

---

#### `editor`

**Ответственность**

* бизнес-логика редактора заметок
* управление собственными документами
* использование файлового хранилища
* проверка доступа пользователя к операциям над документами

**Не ответственность**

* хранение пользовательских учётных данных
* реализация механизма авторизации
* реализация файлового хранилища
* общий аудит платформы

**Данные**

* документы
* состояние документов
* данные редактора

**Зависимости**

* `account`
* `filestorage`
* `access`
* `auditor`

**Потребители**

* `gateway`

---

#### `planner`

**Ответственность**

* бизнес-логика планирования
* управление собственными сущностями планирования
* взаимодействие с календарём в сценариях, где необходимы календарные данные

**Не ответственность**

* управление календарными сущностями
* реализация аудита
* бизнес-логика других модулей

**Данные**

* планы
* задачи и другие сущности планировщика

**Зависимости**

* `calendar`
* `auditor`

**Потребители**

* `gateway`

### 4. Карта владения данными

| Данные                          | Владелец       | Кто использует                 |
|---------------------------------|----------------|--------------------------------|
| Данные профиля пользователя     | `user`         | `user`                         |
| Сессия                          | `session`      | Все компоненты платформы       |
| Данные авторизации пользователя | `auth`         | `gateway`                      |
| Доступы пользователя к модулям  | `access`       | `gateway`                      |
| Уведомления                     | `notification` | `notification`, `calendar`     |
| Данные почтовика                | `mailer`       | `mailer`, `calendar`           |
| Файлы                           | `filestorage`  | `filestorage`, `editor`        |
| События продукта                | `auditor`      | `auditor`                      |
| События календаря               | `calendar`     | `calendar`, `planner`          |
| Данные календаря (чей он и тд)  | `calendar`     | `calendar`                     |
| Документы/Заметки               | `editor`       | `editor`                       |
| Данные запланированных событий  | `planner`      | `planner`                      |
| Реестр модулей                  | `appreg`       | `gateway`, `bff`               |
| Realtime-соединения             | `realtime`     | `фронт`, `продуктовые сервисы` |

### 5. Основные пользовательские сценарии

#### Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant User as user
    Client ->> Frontend: Register
    Frontend ->> Gateway: Register
    Gateway ->> User: Register
    User ->> User: Создать учётные данные
    User ->> Session: CreateSession
    Session -->> User: SessionData
    User -->> Gateway: OK
    Gateway -->> Frontend: OK
    Frontend -->> Client: OK
```

---

#### Вход пользователя

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Auth as auth
    participant UserService as user
    participant Session as session
    User ->> Frontend: Login
    Frontend ->> Gateway: Login
    Gateway ->> Auth: proxy
    Auth ->> UserService: GetUser
    UserService -->> Auth: UserData
    Auth ->> Auth: Проверка кредов
    Auth ->> Session: CreateSession
    Session -->> Auth: SessionData
    Auth -->> Gateway: OK
    Gateway -->> Frontend: OK
```

---

#### Работа с документом в Editor

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Session as session
    participant Access as access
    participant Editor as editor
    participant Storage as filestorage
    participant Auditor as auditor
    User ->> Frontend: Изменить документ
    Frontend ->> Gateway: UpdateDocument
    Gateway ->> Session: CheckSession
    Session -->> Session: GetSession
    Session -->> Gateway: SessionAvailable
    Gateway ->> Access: CheckModuleAvailable
    Access -->> Gateway: Access

    alt Доступ есть
        Gateway ->> Editor: UpdateDocument
        Editor -->> Editor: CheckEditorAccess
        alt Доступ есть
            Editor ->> Storage: Работа с файлом
            Storage -->> Editor: Результат
            Editor -->> Auditor: Зафиксировать действие
            Editor -->> Frontend: OK
        else Доступа нет
            Editor -->> Frontend: Access denied
        end
    else Доступ запрещён
        Gateway -->> Frontend: Access denied
    end
```

---

#### Работа Planner с Calendar

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Planner as planner
    participant Calendar as calendar
    participant Auditor as auditor
    participant Notification as notification
    User ->> Frontend: Запланировать событие
    Frontend ->> Gateway: CratePlannedEvent
    Gateway ->> Planner: CratePlannedEvent
    Planner ->> Calendar: GetCalendarData
    Calendar -->> Planner: Calendar Data
    Planner ->> Planner: CratePlannedEvent
    Planner -->> Notification: PushNotification
    Planner -->> Auditor: AuditEvent
    Planner -->> Gateway: Result
    Gateway -->> Frontend: Result
```
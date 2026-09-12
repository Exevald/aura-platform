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
        Identity[identity]
        Access[access]
        Account[account]
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
    Gateway -.->|proxy| Identity
    Gateway -.->|proxy| Access
    Gateway -.->|proxy| Account
    Gateway -.->|proxy| Notification
    Gateway -.->|proxy| Calendar
    Gateway -.->|proxy| Editor
    Gateway -.->|proxy| Planner
    Access --> Account
    Identity --> Account
    Mailer --> Identity
    Editor --> Account
    Editor --> Filestorage
    Editor --> Access
    Editor .-> Auditor
    Calendar --> Identity
    Calendar -.-> Notification
    Calendar -.-> Mailer
    Calendar -.-> Auditor
    Planner -.-> Calendar
    Planner -.-> Auditor
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

#### `identity`

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

* `account`

**Потребители**

* `gateway`
* `calendar`
* `mailer`
* и другие продуктовые модули

---

#### `account`

**Ответственность**

* хранение данных корпоративного аккаунта пользователей
* управление данными аккаунта
* управление пользователями в аккаунте
* управление доступами в аккаунте

**Не ответственность**

* управление сессиями
* продуктовые данные

**Данные**

* аккаунт пользователя
* общие корпоративние данные данные

**Зависимости**

* нет

**Потребители**

* `identity`
* `gateway`
* `editor`
* `access`
* и другие продуктовые модули

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

| Данные                         | Владелец       | Кто использует                  |
|--------------------------------|----------------|---------------------------------|
| Данные профиля пользователя    | `identity`     | `identity`                      |
| Сессия                         | `gateway`      | Все компоненты платформы        |
| Данные аккаунта                | `account`      | `identity`, `gateway`, `editor` |
| Доступы пользователя           | `access`       | `access`, `editor               |
| Уведомления                    | `notification` | `notification`, `calendar`      |
| Данные почтовика               | `mailer`       | `mailer`, `calendar`            |
| Файлы                          | `filestorage`  | `filestorage`, `editor`         |
| События продукта               | `auditor`      | `auditor`                       |
| События календаря              | `calendar`     | `calendar`, `planner`           |
| Данные календаря (чей он и тд) | `calendar`     | `calendar`                      |
| Документы/Заметки              | `editor`       | `editor`                        |
| Данные запланированных событий | `planner`      | `planner`                       |

### 5. Основные пользовательские сценарии

#### 5.1. Регистрация пользователя

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Identity as identity
    participant Account as account
    User ->> Frontend: Регистрационные данные
    Frontend ->> Gateway: Register
    Gateway ->> Identity: Register
    Identity ->> Account: Создать аккаунт
    Account -->> Identity: Account
    Identity ->> Identity: Создать учётные данные
    Identity -->> Gateway: Пользователь зарегистрирован
    Gateway -->> Frontend: Результат
    Frontend -->> User: Регистрация завершена
```

---

#### 5.2. Вход пользователя

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Identity as identity
    User ->> Frontend: Логин и пароль
    Frontend ->> Gateway: Login
    Gateway ->> Identity: Authenticate
    Identity ->> Identity: Проверить учётные данные
    Identity -->> Gateway: ОК
    Gateway ->> Gateway: Создать сессию
    Gateway -->> Frontend: Пользователь аутентифицирован
    Frontend -->> User: Вход выполнен
```

---

#### 5.3. Получение данных корпоративного аккаунта

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Account as account
    User ->> Frontend: Открыть профиль
    Frontend ->> Gateway: Получить данные корпоративного аккаунта
    Gateway ->> Account: GetAccount
    Account -->> Gateway: Account
    Gateway -->> Frontend: Account
    Frontend -->> User: Показать данные
```

---

#### 5.4. Работа с документом в Editor

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Editor as editor
    participant Access as access
    participant Storage as filestorage
    participant Auditor as auditor
    User ->> Frontend: Изменить документ
    Frontend ->> Gateway: UpdateDocument
    Gateway ->> Editor: UpdateDocument
    Editor ->> Access: CheckAccess
    Access -->> Editor: Allow / Deny

    alt Доступ разрешён
        Editor ->> Storage: Работа с файлом
        Storage -->> Editor: Результат
        Editor -->> Auditor: Зафиксировать действие
        Editor -->> Frontend: OK
    else Доступ запрещён
        Editor -->> Frontend: Access denied
    end
```

---

#### 5.5. Работа с Calendar

```mermaid
sequenceDiagram
    actor User as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Calendar as calendar
    participant Identity as identity
    participant Notification as notification
    participant Mailer as mailer
    participant Auditor as auditor
    User ->> Frontend: Создать событие
    Frontend ->> Gateway: CreateEvent
    Gateway ->> Calendar: CreateEvent
    Calendar ->> Identity: GetUserData
    Identity -->> Calendar: UserData
    Calendar ->> Calendar: CreateCalendarEvent
    Calendar -->> Notification: PushNotification
    Calendar -->> Mailer: SendMail
    Calendar -->> Auditor: AuditEvent
    Calendar -->> Frontend: Событие создано
```

---

#### 5.6. Работа Planner с Calendar

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
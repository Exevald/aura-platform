# Сценарии работы системы

## Назначение модулей группе пользователей

```mermaid
sequenceDiagram
    actor Owner as Владелец группы
    participant Access
    participant UserGroup
    participant Auditor
    Owner->>Access: PUT groups/{groupID}/modules(moduleIDs)
    Access->>UserGroup: GetOwnedGroupMembers
    alt Группа принадлежит владельцу
        UserGroup-->>Access: member UUIDs
        Access->>Access: добавить владельца к получателям и атомарно заменить назначения
        Access-->>Owner: итоговый набор модулей
        Access-->>Auditor: GroupModuleAccessGranted / Revoked
    else Чужая группа
        Access-->>Owner: 404 без частичных изменений
    end
```

## Приглашение в рабочее пространство

```mermaid
sequenceDiagram
    actor Owner as Владелец
    actor Invitee as Участник
    participant Gateway
    participant Workspace
    Owner->>Gateway: Создать invitation(email/userID, role)
    Gateway->>Workspace: session + workspace membership check
    Workspace->>Workspace: Сохранить token и срок действия
    Workspace-->>Owner: pending invitation
    Invitee->>Gateway: Принять token
    Gateway->>Workspace: authenticated user + token
    Workspace->>Workspace: Атомарно создать membership и удалить invitation
    Workspace-->>Invitee: workspace с назначенной ролью
```

## Напоминание календаря

```mermaid
sequenceDiagram
    actor User
    participant Calendar
    participant Notification
    participant Realtime
    participant Mailer
    User->>Calendar: Создать event с attendees и reminders
    Calendar->>Calendar: Event + reminder deliveries + outbox в одной транзакции
    loop По расписанию
        Calendar->>Notification: CreateNotification(idempotencyKey)
        Notification->>Notification: Inbox + outbox
        Notification-->>Realtime: Publish user/workspace event
        opt Разрешён email-канал
            Notification-->>Mailer: SendMail(idempotencyKey)
        end
        Calendar->>Calendar: Завершить или перенести recurring reminder
    end
```

## Задача с синхронизацией в календарь

```mermaid
sequenceDiagram
    actor User
    participant Planner
    participant Calendar
    User->>Planner: Создать/изменить задачу с dueAt и calendarID
    Planner->>Planner: Сохранить задачу и outbox
    Planner->>Calendar: UpsertTaskEvent(taskID, dueAt)
    alt Calendar недоступен при создании
        Planner->>Planner: Компенсировать создание задачи
        Planner-->>User: 503
    else Успех
        Calendar-->>Planner: eventID
        Planner->>Planner: Сохранить calendarEventID
        Planner-->>User: task
    end
```

## Вложение документа

```mermaid
sequenceDiagram
    actor User
    participant Gateway
    participant Editor
    participant Filestorage
    participant MinIO
    User->>Gateway: multipart attachment + X-Workspace-ID
    Gateway->>Gateway: session, workspace и module access
    Gateway->>Editor: upload
    Editor->>Editor: Проверить роль owner/editor документа
    Editor->>Filestorage: streaming Upload
    Filestorage->>MinIO: object
    Filestorage->>Filestorage: metadata + outbox
    Editor->>Editor: Привязать fileID к documentID
    Editor-->>User: attachment metadata
```

`access` — единственный владелец назначений модулей группам. Каждый источник
«группа–модуль» хранится отдельно, поэтому отзыв одного источника не удаляет доступ,
который пользователь продолжает получать через другую группу. Изменения состава
группы добавляют или удаляют гранты всех её активных модулей, а владелец получает
их автоматически, даже если в группе ещё нет участников. Права на отдельные
ресурсы по-прежнему принадлежат соответствующим продуктовым сервисам.

## Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Registerer as registerer
    participant RegistererDB as registererDB
    participant Auth as auth
    participant AuthDB as authDB
    participant Session as session
    participant SessionDB as sessionDB
    participant User as user
    participant UserDB as userDB
    participant Broker
    participant Auditor as auditor
    participant AuditDB as auditDB
    Client ->> Frontend: Register
    Frontend ->> Gateway: /api/public/v1/registerer/register
    Note over Gateway: Создать requestID
    Gateway ->> Registerer: Register(login, password, profile)
    Registerer ->> RegistererDB: INSERT transaction(preparing, lease)
    Note over Registerer,UserDB: transactionID и userID создаёт registerer
    Registerer ->> Auth: PrepareCredentialRegistration(transactionID, userID, login, password)
    Auth ->> AuthDB: XA PREPARE credential
    AuthDB -->> Auth: PREPARED
    Auth -->> Registerer: credentialID
    Registerer ->> User: PrepareUserRegistration(transactionID, userID, profile)
        User ->> UserDB: XA PREPARE user + UserCreated outbox
        UserDB -->> User: PREPARED
    User -->> Registerer: PREPARED
    alt Все участники подготовлены
        Registerer ->> RegistererDB: сохранить COMMIT + credentialID
        Registerer ->> Auth: CommitCredentialRegistration
        Auth ->> AuthDB: XA COMMIT auth
        Registerer ->> User: CommitUserRegistration
        User ->> UserDB: XA COMMIT user
        Registerer ->> RegistererDB: state=committed + UserRegistered outbox
        Note over Registerer,User: Недоставленный COMMIT повторяет recovery worker
    else Ошибка prepare
        Registerer ->> RegistererDB: сохранить ABORT
        Registerer ->> Auth: RollbackCredentialRegistration
        Auth ->> AuthDB: XA ROLLBACK auth
        Registerer ->> User: RollbackUserRegistration
        User ->> UserDB: XA ROLLBACK user
        Registerer -->> Gateway: Registration unavailable/conflict
    end
    alt Session доступен после auth commit
            Registerer ->> Session: CreateSession(user_id, rememberMe)
            Session ->> SessionDB: INSERT session
            SessionDB -->> Session: OK
            Session -->> Broker: SessionCreated
            Session -->> Registerer: session_id, expires_at
    else Session не доступен
        Note over Registerer,Session: Регистрация committed, authenticated=false
    end
    Registerer -->> Broker: UserRegistered из committed outbox
    Registerer -->> Gateway: RegisterResult(user_id)
    Gateway -->> Frontend: 201
    Frontend -->> Client: Registration successful

    alt Broker доступен
            Auditor ->> Auditor: HandleEvent
            Auditor ->> AuditDB: INSERT audit event
            AuditDB -->> Auditor: OK
    else Broker недоступен
        Note over Registerer, Broker: Outbox relay повторяет публикацию
    end
```

## Логинация пользователя

```mermaid
sequenceDiagram
    actor Client as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Session as session
    participant SessionDB as sessionDB
    participant Auth as auth
    participant AuthDB as authDB
    participant Broker
    participant Auditor as auditor
    participant AuditDB as auditDB
    Client ->> Frontend: Login(login, password)
    Frontend ->> Gateway: /login(login, password, rememberMe)
    Note over Gateway: Создать request_id
    Gateway ->> Session: CheckSession(session_id)

    alt Сессия доступна
        Session ->> Gateway: Session available
        Gateway ->> Frontend: Переход на главную страницу
    else Сессия отсутствует / истекла
        Gateway ->> Auth: Login(login, password)
        Auth ->> AuthDB: FindCredential(login)

        alt Пользователь не найден
            AuthDB -->> Auth: not found
            Auth -->> Gateway: InvalidCredentials
            Gateway -->> Frontend: 401 Unauthorized
            Frontend -->> Client: Неверный login или password

        else Credential найден
            AuthDB -->> Auth: credential + user_id + password_hash
            Auth ->> Auth: VerifyPassword(password, hash)

            alt Password неверный
                Auth -->> Broker: InvalidPasswordGiven
                Auth -->> Gateway: InvalidCredentials
                Gateway -->> Frontend: 401
                Frontend -->> Client: Неверный login или password

            else Password верный
                Auth ->> Session: CreateSession(user_id, rememberMe)
                Session ->> SessionDB: INSERT session
                SessionDB -->> Session: OK
                Session -->> Auth: session_id, expires_at
                Auth -->> Gateway: LoginResult(session_id)
                Auth -->> Broker: UserLogin
                Gateway -->> Frontend: Set-Cookie / token
                Frontend -->> Client: Login successful
            end
        end
    end
    alt Broker доступен
        Auditor ->> Auditor: HandleEvent
        Auditor ->> AuditDB: INSERT audit event
        AuditDB -->> Auditor: OK
    else Broker недоступен
        Note over Auth, Broker: Логируем ошибку и ретраим обработку события
    end
```

## Logout пользователя

```mermaid
sequenceDiagram
    actor Client as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Session as session
    participant SessionDB as sessionDB
    participant Broker
    participant Auditor as auditor
    participant AuditDB as auditDB
    Client ->> Frontend: Logout
    Frontend ->> Gateway: /logout
    Note over Gateway: Создать request_id
    Gateway ->> Session: RevokeSession(session_id)
    Session ->> SessionDB: GetSession(session_id)

    alt Сессия отсутствует / уже завершена
        SessionDB -->> Session: not found / inactive
        Session -->> Gateway: Not found
        Gateway -->> Frontend: 400
        Frontend -->> Client: Session not found

    else Сессия существует
        SessionDB -->> Session: session data
        Session ->> SessionDB: RevokeSession(session_id)
        SessionDB -->> Session: OK
        Session -->> Broker: SessionRevoked
        Session -->> Gateway: OK
        Gateway -->> Frontend: 204
        Frontend -->> Client: Logout successful
    end
    alt Broker доступен
        Auditor ->> Auditor: HandleEvent
        Auditor ->> AuditDB: INSERT audit event
        AuditDB -->> Auditor: OK
    else Broker недоступен
        Note over Auth, Broker: Логируем ошибку и ретраим обработку события
    end
```

## Смена пароля

```mermaid
sequenceDiagram
    actor Client as Пользователь
    participant Frontend
    participant Gateway as gateway
    participant Session as session
    participant SessionDB as sessionDB
    participant Auth as auth
    participant AuthDB as authDB
    participant Broker
    participant Auditor as auditor
    participant AuditDB as auditDB
    Client ->> Frontend: ChangePassword(currentPassword, newPassword)
    Frontend ->> Gateway: /change-password
    Note over Gateway: Создать request_id
    Gateway ->> Session: CheckSession(session_id)
    Session ->> SessionDB: FindSession(session_id)

    alt Сессия отсутствует / истекла
        SessionDB -->> Session: not found / inactive
        Session -->> Gateway: InvalidSession
        Gateway -->> Frontend: 401
        Frontend -->> Client: Требуется login

    else Сессия доступна
        SessionDB -->> Session: user_id, status, expires_at
        Session -->> Gateway: Authenticated(user_id)
        Gateway ->> Auth: ChangePassword(user_id, currentPassword, newPassword)
        Auth ->> AuthDB: FindCredential(user_id)
        AuthDB -->> Auth: password_hash
        Auth ->> Auth: VerifyPassword(currentPassword, password_hash)

        alt Текущий пароль неверный
            Auth -->> Gateway: InvalidCredentials
            Gateway -->> Frontend: 400
            Frontend -->> Client: Неверный текущий пароль

        else Текущий пароль верный
            Auth ->> Auth: ValidatePassword(newPassword)

            alt Новый пароль не соответствует требованиям
                Auth -->> Gateway: InvalidPassword
                Gateway -->> Frontend: 400
                Frontend -->> Client: Новый пароль не соответствует требованиям

            else Новый пароль корректный
                Auth ->> AuthDB: UpdateCredential(user_id, new_password_hash)
                AuthDB -->> Auth: OK
                Auth ->> Session: RevokeSessions
                Session ->> SessionDB: DELETE session 
                SessionDB -->> Session: OK
                Session -->> Broker: SessionRevoked
                Session -->> Auth: Revoked
                Auth -->> Broker: PasswordChanged
                Auth -->> Gateway: PasswordChanged
                Gateway -->> Frontend: 204
                Frontend -->> Client: Password changed
            end
        end
    end
    alt Broker доступен
        Auditor ->> Auditor: HandleEvent
        Auditor ->> AuditDB: INSERT audit event
        AuditDB -->> Auditor: OK
    else Broker недоступен
        Note over Auth, Broker: Логируем ошибку и ретраим обработку события
    end
```

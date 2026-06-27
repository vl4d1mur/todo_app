# Todo App — Микросервисное приложение

REST API приложение для управления задачами и заметками с микросервисной архитектурой, событийной коммуникацией и email+telegram-уведомлениями.

## Архитектура

Проект состоит из трёх сервисов: `auth_service`, `task_service`, `notifier_service`. Каждый имеет свою БД, общая инфраструктура — NATS как шина событий и MailHog для SMTP.

### Компонентная диаграмма

```mermaid
graph TB
    Client[Client]

    subgraph Gateway
        Nginx[nginx :8080]
    end

    subgraph AuthDomain[Auth Domain]
        Auth[auth_service<br/>HTTP :8091<br/>gRPC :50051]
        PgAuth[(postgres-auth)]
        RedisAuth[(redis-auth)]
    end

    subgraph TaskDomain[Task Domain]
        Task[task_service<br/>HTTP :8090]
        PgTasks[(postgres-tasks)]
        RedisTasks[(redis-tasks)]
        Mongo[(mongo)]
    end

    subgraph NotifierDomain[Notifier Domain]
        Notifier[notifier_service<br/>HTTP :8092]
        MongoNotif[(mongo)]
        MailHog[mailhog<br/>SMTP :1025]
        TgBot[Telegram Bot API]
    end

    subgraph Broker[Event Broker]
        NATS[nats :4222]
    end

    Client -->|HTTP| Nginx
    Nginx -->|/api/register<br/>/api/login<br/>/api/refresh<br/>/api/logout<br/>/api/profile| Auth
    Nginx -->|/api/tasks/*<br/>/api/notes/*| Task
    Nginx -->|/api/notifications| Notifier

    Auth --> PgAuth
    Auth --> RedisAuth

    Task --> PgTasks
    Task --> RedisTasks
    Task --> Mongo
    Task -.->|ValidateToken| Auth
    Task -->|publish events| NATS

    Notifier --> MongoNotif
    Notifier -.->|GetUserEmail| Auth
    Notifier -->|SMTP| MailHog
    Notifier -->|HTTPS| TgBot
    NATS -->|subscribe task-events| Notifier
```

### Sequence — прецедент изменение статуса задачи

```mermaid
sequenceDiagram
    actor Client
    participant Nginx as nginx
    participant Task as task_service
    participant Auth as auth_service
    participant Postgres as postgres-tasks
    participant Redis as redis-tasks
    participant NATS
    participant Notifier as notifier_service
    participant MongoN as mongo (notifier)
    participant SMTP as mailhog

    Client->>Nginx: PATCH /api/tasks/{id}<br/>Authorization: Bearer
    Nginx->>Task: forward request

    Task->>Auth: gRPC ValidateToken
    Auth-->>Task: user_id

    Task->>Postgres: UPDATE tasks SET status
    Postgres-->>Task: ok

    Task->>Redis: DEL cache:tasks:user_id
    Task->>NATS: publish task.status_changed

    Task-->>Nginx: 200 OK
    Nginx-->>Client: 200 OK

    NATS->>Notifier: event received
    Notifier->>Auth: gRPC GetUserEmail
    Auth-->>Notifier: email

    Notifier->>MongoN: INSERT notification (pending)
    Notifier->>SMTP: send email

    alt email sent successfully
        SMTP-->>Notifier: ok
        Notifier->>MongoN: UPDATE status = sent
    else email failed
        SMTP-->>Notifier: error
        Notifier->>MongoN: UPDATE status = failed
    end
```

### Sequence — прецедент уведомление о приближении дедлайна

```mermaid
sequenceDiagram
    actor Client
    participant Task as task_service
    participant NATS
    participant Notifier as notifier_service
    participant MongoN as mongo (notifier)
    participant Auth as auth_service
    participant SMTP as mailhog

    Note over Client,Task: При создании или изменении задачи
    Client->>Task: POST/PATCH /api/tasks
    Task->>NATS: publish task.created / task.updated<br/>{title, deadline}
    NATS->>Notifier: event received
    Notifier->>MongoN: UPSERT task_deadlines

    Note over Notifier: Каждый час cron проверяет<br/>приближающиеся дедлайны

    loop Every hour
        Notifier->>MongoN: SELECT deadlines<br/>WHERE deadline < now + 24h<br/>AND notified = false
        MongoN-->>Notifier: pending deadlines

        loop For each deadline
        Notifier->>Auth: gRPC GetUserContacts
        Auth-->>Notifier: email, chat_id, has_telegram
            
        par Send email
        Notifier->>SMTP: send email
        and Send telegram (if linked)
        Notifier->>TgBot: send message
        end
            SMTP-->>Notifier: ok
            Notifier->>MongoN: UPDATE notification status = sent
            Notifier->>MongoN: UPDATE deadline notified = true
        end
    end
```

## Как поднять локально

Требуется только Docker.

```bash
docker compose up --build
```

Доступные точки:

- API через nginx — `http://localhost:8080`
- MailHog UI — `http://localhost:8025`
- NATS мониторинг — `http://localhost:8222`

Миграции применяются автоматически отдельными контейнерами через goose.

Остановка с сохранением данных: `docker compose down`. Полная очистка: `docker compose down -v`.

## Переменные окружения

Каждый сервис имеет свой `.env.example` в `<service>/config/`.

### auth_service

```dotenv
SERVER_PORT=:8091
GRPC_PORT=:50051
POSTGRES_DSN=postgres://postgres:root@postgres-auth:5432/users
JWT_SECRET=your-secret-here
JWT_REFRESH_SECRET=your-refresh-secret-here
REDIS_ADDR=redis-auth:6379
APP_ENV=development
```

### task_service

```dotenv
SERVER_PORT=:8090
POSTGRES_DSN=postgres://postgres:root@postgres-tasks:5432/tasks
REDIS_ADDR=redis-tasks:6379
MONGO_URI=mongodb://mongo:27017
MONGO_DB_NAME=todo_notes
MONGO_DB_COLLECTION=notes
NATS_URL=nats://nats:4222
AUTH_GRPC_ADDR=auth_service:50051
APP_ENV=development
```

### notifier_service

```dotenv
SERVER_PORT=:8092
MONGO_URI=mongodb://mongo:27017
MONGO_DB_NAME=notifier_db
NATS_URL=nats://nats:4222
AUTH_GRPC_ADDR=auth_service:50051
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_FROM=noreply@todo.local
TELEGRAM_BOT_TOKEN=(token)
APP_ENV=development
```

## Примеры запросов

Все запросы идут через nginx на порт `8080`. Защищённые эндпоинты требуют заголовок `Authorization: Bearer <TOKEN>`.

### Auth

```bash
# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# Логин
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Профиль
curl -X GET http://localhost:8080/api/profile -H "Authorization: Bearer <TOKEN>"

# Refresh
curl -X POST http://localhost:8080/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'

# Logout
curl -X POST http://localhost:8080/api/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

### Tasks

```bash
# Создать
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Buy milk","status":"todo","priority":2,"deadline":"2026-06-30T12:00:00Z"}'

# Список с пагинацией и фильтром
curl -X GET "http://localhost:8080/api/tasks?page=1&limit=10&status=todo" \
  -H "Authorization: Bearer <TOKEN>"

# По ID
curl -X GET http://localhost:8080/api/tasks/<TASK_ID> -H "Authorization: Bearer <TOKEN>"

# Обновить
curl -X PATCH http://localhost:8080/api/tasks/<TASK_ID> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"status":"in_progress"}'

# Удалить
curl -X DELETE http://localhost:8080/api/tasks/<TASK_ID> -H "Authorization: Bearer <TOKEN>"
```

### Notes

```bash
# Добавить
curl -X POST http://localhost:8080/api/tasks/<TASK_ID>/notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"text":"Note","meta":{"color":"red"}}'

# Список
curl -X GET http://localhost:8080/api/tasks/<TASK_ID>/notes -H "Authorization: Bearer <TOKEN>"

# Удалить
curl -X DELETE http://localhost:8080/api/notes/<NOTE_ID> -H "Authorization: Bearer <TOKEN>"
```

### Notifications

```bash
curl -X GET "http://localhost:8080/api/notifications?page=1&limit=10" \
  -H "Authorization: Bearer <TOKEN>"
```

### Health

```bash
curl http://localhost:8080/healthz/auth
curl http://localhost:8080/readyz/auth
curl http://localhost:8080/healthz/tasks
curl http://localhost:8080/readyz/tasks
curl http://localhost:8080/healthz/notifier
curl http://localhost:8080/readyz/notifier
```
## CI/CD

Pipeline в GitHub Actions запускается на push и PR в `main`:

- **Lint** — golangci-lint для каждого сервиса
- **Unit-тесты** — `go test`
- **Integration-тесты** — с testcontainers (поднимает Postgres, Redis, MongoDB, NATS, MailHog)
- **Docker build** — проверка что Dockerfile собираются

Все 4 типа job выполняются параллельно через matrix strategy для всех трёх сервисов.

## Telegram интеграция

Notifier поддерживает доставку уведомлений о дедлайнах через Telegram бота.

### Привязка пользователя

1. Получить код привязки:
```bash
   curl -X POST http://localhost:8080/api/profile/telegram/code \
     -H "Authorization: Bearer <TOKEN>"
```
2. Открыть своего бота в Telegram, отправить: /start (code)
3. Бот ответит подтверждением, после чего уведомления о приближении дедлайнов начнут приходить в Telegram.
Уведомления о смене статуса по-прежнему идут только в email.

## Выбор библиотек

- **gorilla/mux** — HTTP роутинг. Я просто немного фрик
- **rs/zerolog** — структурированное логирование, быстрый и просто нравится.
- **jackc/pgx/v5** — нативный Postgres драйвер. 
- Мужчину спрашивают:
  - Что вы любите больше всего?
  - Писать SQL запросы! Писал бы 24 часа в сутки, не отрываясь! Я от одного вида запросов трясусь, могу писать SQL бесконечно!
  - Понятно, а что еще любите?
  - В дурке лежать люблю! Лежал бы круглыми сутками в смерительной рубашке, в тихой палате, что б никто не отвлекал, лиж бы ноутбук был!
  - Эээ... А ноутбук зачем?
  - SQL запросы писать люблю!
- **mongo-driver/v2** — официальный драйвер MongoDB актуальной версии.
- **pressly/goose** — миграции в SQL файлах с `-- +goose Up/Down` секциями.
- **google.golang.org/grpc** — стандартный gRPC, настроены keepalive и retry policy через serviceConfig.
- **nats-io/nats.go** — лёгкий брокер для микросервисов, без JetStream (достаточно at-most-once для текущих сценариев). 
- **golang-jwt/jwt/v5** — JWT с подписью для обоих токенов (access и refresh).
- **golang.org/x/crypto/bcrypt** — стандарт для хэширования паролей.
- **gopkg.in/gomail.v2** — обёртка над `net/smtp`.
- **google/uuid** — UUID как идентификаторы во всех сервисах для изоляции данных.
- **go-playground/validator/v10** — декларативная валидация через теги.
- **mailhog/mailhog** — тестовый SMTP сервер с веб-интерфейсом.
- **testcontainers-go** — интеграционные тесты на реальных Postgres, Redis, MongoDB, NATS, MailHog.
- **stretchr/testify** — удобные assertions в тестах.

## Что не успели / планы по улучшению

### Запланировано

- **Observability через Prometheus + Grafana [WIP]** — метрики каждого сервиса и дашборды.

### Технический долг

- **Пагинация в task_service через SQL** — сейчас фильтрация и нарезка списка задач происходит в памяти Go(так вышло). Перевести на `LIMIT/OFFSET` для масштабируемости.
- **JWT blacklist через Redis** — инвалидация access токенов при logout. Архитектура для этого уже подготовлена. [WIP]

### Альтернативные библиотеки на будущее

- **go.uber.org/fx** для DI — убрать ручное создание зависимостей в `main.go`.
- **NATS JetStream** — at-least-once доставка с персистентностью сообщений.

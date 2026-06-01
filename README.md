# Todo App — Этап 1

## Стек

| Компонент  | Назначение                        |
|------------|-----------------------------------|
| Go         | Язык                              |
| PostgreSQL | Пользователи, задачи, сессии      |
| MongoDB    | Заметки к задачам                 |
| Redis      | Кэш списка задач и профиля        |
| NATS       | Брокер событий по задачам         |
| Docker     | Запуск всего окружения            |

---

## API

### Auth

```bash
# Регистрация
curl -X POST http://localhost:8090/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# Логин
curl -X POST http://localhost:8090/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Обновить токены
curl -X POST http://localhost:8090/api/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'

# Логаут
curl -X POST http://localhost:8090/api/logout \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'

# Профиль
curl -X GET http://localhost:8090/api/profile \
  -H "Authorization: Bearer <TOKEN>"
```

### Tasks

```bash
# Создать задачу
curl -X POST http://localhost:8090/api/tasks \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Купить молоко","description":"Обязательно 3.5%","status":"todo","priority":2,"deadline":"2026-06-01T12:00:00Z"}'

# Список задач (пагинация + фильтр по статусу)
curl -X GET "http://localhost:8090/api/tasks?page=1&limit=10&status=todo" \
  -H "Authorization: Bearer <TOKEN>"

# Получить задачу с заметками
curl -X GET http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Authorization: Bearer <TOKEN>"

# Обновить задачу
curl -X PATCH http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk","status":"in_progress"}'

# Удалить задачу
curl -X DELETE http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Authorization: Bearer <TOKEN>"
```

### Notes

```bash
# Создать заметку
curl -X POST http://localhost:8090/api/tasks/<TASK_ID>/notes \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"text":"Тестовая заметка","meta":{"color":"red","pinned":true}}'

# Список заметок задачи
curl -X GET http://localhost:8090/api/tasks/<TASK_ID>/notes \
  -H "Authorization: Bearer <TOKEN>"

# Удалить заметку
curl -X DELETE http://localhost:8090/api/notes/<NOTE_ID> \
  -H "Authorization: Bearer <TOKEN>"
```

### Health

```bash
# Liveness
curl http://localhost:8090/healthz

# Readiness
curl http://localhost:8090/readyz
```
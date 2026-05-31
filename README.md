# Примеры PowerShell запросов

## Auth

# Регистрация
Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/register -Headers @{"Content-Type"="application/json"} -Body '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# Логин (token)
Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/login -Headers @{"Content-Type"="application/json"} -Body '{"email":"john@example.com","password":"password123"}'

# Профиль пользователя
Invoke-RestMethod -Method GET -Uri http://localhost:8090/api/profile -Headers @{"Authorization"="Bearer <TOKEN>"}

## Tasks

# Создать задачу
Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/tasks -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"title":"Купить молоко","description":"Обязательно 3.5%","status":"todo","priority":2,"deadline":"2026-06-01T12:00:00Z"}'

# Получить все задачи пользователя
Invoke-RestMethod -Method GET -Uri http://localhost:8090/api/tasks -Headers @{"Authorization"="Bearer <TOKEN>"}

# Получить задачу по ID
Invoke-RestMethod -Method GET -Uri http://localhost:8090/api/tasks/<TASK_ID> -Headers @{"Authorization"="Bearer <TOKEN>"}

# Обновить задачу
Invoke-RestMethod -Method PATCH -Uri http://localhost:8090/api/tasks/<TASK_ID> -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"title":"Buy milk22","status":"in_progress"}'

# Удалить задачу
Invoke-RestMethod -Method DELETE -Uri http://localhost:8090/api/tasks/<TASK_ID> -Headers @{"Authorization"="Bearer <TOKEN>"}


## Notes

# Создать заметку к задаче
Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/tasks/<TASK_ID>/notes -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"text":"Это тестовая заметка","meta":{"color":"red","pinned":true}}'

# Получить все заметки задачи
Invoke-RestMethod -Method GET -Uri http://localhost:8090/api/tasks/<TASK_ID>/notes -Headers @{"Authorization"="Bearer <TOKEN>"}

# Удалить заметку
Invoke-RestMethod -Method DELETE -Uri http://localhost:8090/api/notes/<NOTE_ID> -Headers @{"Authorization"="Bearer <TOKEN>"}

---

# Примеры Curl запросов

## Auth

# Регистрация
curl -X POST http://localhost:8090/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123"}'

# Логин
curl -X POST http://localhost:8090/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'

# Профиль пользователя
curl -X GET http://localhost:8090/api/profile \
  -H "Authorization: Bearer <TOKEN>"

# Tasks

# Создать задачу
curl -X POST http://localhost:8090/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Купить молоко","description":"Обязательно 3.5%","status":"todo","priority":2,"deadline":"2026-06-01T12:00:00Z"}'

# Получить все задачи
curl -X GET http://localhost:8090/api/tasks \
  -H "Authorization: Bearer <TOKEN>"

# Получить задачу по ID
curl -X GET http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Authorization: Bearer <TOKEN>"

# Обновить задачу
curl -X PATCH http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Buy milk22","status":"in_progress"}'

# Удалить задачу
curl -X DELETE http://localhost:8090/api/tasks/<TASK_ID> \
  -H "Authorization: Bearer <TOKEN>"

# Notes

# Создать заметку к задаче
curl -X POST http://localhost:8090/api/tasks/<TASK_ID>/notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"text":"Это тестовая заметка","meta":{"color":"red","pinned":true}}'

# Получить все заметки задачи
curl -X GET http://localhost:8090/api/tasks/<TASK_ID>/notes \
  -H "Authorization: Bearer <TOKEN>"

# Удалить заметку
curl -X DELETE http://localhost:8090/api/notes/<NOTE_ID> \
  -H "Authorization: Bearer <TOKEN>"

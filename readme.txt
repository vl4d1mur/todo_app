Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/tasks -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"title":"Купить молоко","description":"Обязательно 3.5%","status":"todo","priority":2,"deadline":"2026-06-01T12:00:00Z"}' // Созадть задачу

Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/login -Headers @{"Content-Type"="application/json"} -Body '{"email":"john@example.com","password":"password123"}' // Залогинится

Invoke-RestMethod -Method POST -Uri http://localhost:8090/api/register -Headers @{"Content-Type"="application/json"} -Body '{"name":"John Doe","email":"john@example.com","password":"password123"}' //Регистрация

Invoke-RestMethod -Method GET -Uri "http://localhost:8080/api/tasks" -Headers @{"Authorization" = "Bearer <TOKEN>"} // Получить все задачи пользователя

Invoke-RestMethod -Method PATCH -Uri http://localhost:8090/api/tasks/<TASK ID> -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"title":"Buy milk22","status":"in_progress"}' // Редактирование задачи

Invoke-RestMethod -Uri "http://localhost:8090/api/tasks/<TASK ID>/notes" -Method POST -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer <TOKEN>"} -Body '{"text":"Это тестовая заметка"}' // Создать заметку к задаче

Invoke-RestMethod -Uri "http://localhost:8090/api/tasks/<TASK ID>/notes" -Method GET -Headers @{ "Authorization" = "Bearer <TOKEN>" } // Получить все заметки задачи


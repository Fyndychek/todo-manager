📋 TODO Manager

Простой и надёжный REST API-сервис для управления задачами с автоматическим CI/CD.
todo-manager — это backend-сервис на Go, предоставляющий REST API для работы со списком задач (TODO). Он поддерживает создание, чтение, удаление и получение статистики по задачам. Всё это упаковано в Docker-образ и автоматически деплоится на сервер при каждом пуше в основную ветку.

🚀 Возможности

    ✅ Создание задачи (с автоматическим ID и временем создания)

    📋 Получение списка всех задач

    ❌ Удаление задачи по ID

    📊 Получение статистики (всего, выполнено, осталось)

    💚 Health‑check для мониторинга

    📦 Встроенная поддержка метрик (Prometheus-совместимый /metrics)

    🔄 Автоматический деплой через GitHub Actions

🛠 Стек технологий

    Язык	Go 1.22+
    Веб-фреймворк	Стандартный net/http
    Хранилище	In‑memory (легко заменить на БД)
    Тестирование	Встроенный testing + httptest
    Контейнеризация	Docker (многоступенчатая сборка)
    CI/CD	GitHub Actions

⚡ Быстрый старт

Локальный запуск (без Docker)

    1.Клонируйте репозиторий:

    git clone https://github.com/your-username/todo-manager.git
    cd todo-manager
    
    2.Установите зависимости:

    go mod download

    3.Запустите сервер:

    go run cmd/server/main.go

    Сервер будет доступен на http://localhost:8080.

Запуск через Docker

    docker build -t todo-manager .
    docker run -p 8080:8080 todo-manager

📌 API Эндпоинты
    
    GET	/health	Проверка работоспособности
    GET	/version	Версия сервиса    
    GET	/todos	Получить список всех задач
    POST	/todos	Создать новую задачу
    DELETE	/todos/{id}	Удалить задачу по ID
    GET	/todos/stats	Получить статистику
    GET	/metrics	Метрики для Prometheus
    
📝 Примеры запросов

Создать задачу:

    curl -X PATCH http://localhost:8081/todos/1 
      -H "Content-Type: application/json" 
      -d "{\"title\":\"buy bread\",\"completed\":true}"

Ответ на создание задачи:

    {
      "id": 1,
      "title": "buy bread",
      "completed": false,
      "created_at": "2025-03-16T12:34:56Z"
    }

Получить все задачи:

    curl http://localhost:8080/todos

Удалить задачу (ID=1):

    curl -X DELETE http://localhost:8080/todos/1

Получить статистику:

    curl http://localhost:8080/todos/stats

ответ на запрос статистики:

    {
      "total": 5,
      "completed": 2,
      "pending": 3
    }

изменить содержание или статус задачи:

    curl -X PATCH http://localhost:8080/todos/1 
      -H "Content-Type: application/json" 
      -d "{\"title\":\"buy bread\",\"completed\":\true\}"

📁 Структура проекта


    .
    ├── cmd/
    |   └── sandbox/ 
    │   └── server/            # Точка входа
    │       └── main.go
    ├── internal/
    │   ├── config/            # Конфигурация (порт и т.д.)
    │   ├── handlers/          # HTTP-обработчики и тесты
    │   ├── models/            # Модели данных
    │   └── storage/           # Интерфейс хранилища + реализация (in-memory)
    ├── Dockerfile
    ├── docker-compose.yml
    ├── go.mod
    ├── go.sum
    └── .github/
        └── workflows/
            └── ci.yml         # Пайплайн GitHub Actions

🧪 Тестирование

Запустите все тесты:

    go test ./...

Для покрытия кода:

    go test -cover ./...

🔄 CI/CD

При каждом пуше в ветку main или develop GitHub Actions автоматически выполняет:

    Линтинг (golangci-lint)

    Юнит-тесты

    Сборку Docker-образа

    Публикацию образа в GitHub Container Registry (или Docker Hub)

Это гарантирует, что код всегда находится в работоспособном состоянии и готов к деплою.


Разработано с ❤️ для упрощения жизни разработчиков.

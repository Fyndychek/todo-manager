📋 TODO Manager

Простой и надёжный REST API-сервис для управления задачами с поддержкой авторизации, персистентного хранения и готовым веб-интерфейсом.

todo-manager — это полнофункциональный бэкенд на Go, предоставляющий REST API для работы со списком задач (TODO), а также одностраничное веб-приложение для удобного управления. Сервис поддерживает JWT-аутентификацию, хранение данных в SQLite, фильтрацию, сортировку и получение статистики. Всё упаковано в Docker-образ и автоматически деплоится при пуше в основную ветку.
🚀 Возможности

    🔐 Аутентификация и авторизация
    Регистрация и вход по логину/паролю с выдачей JWT-токена. Все эндпоинты (кроме /register, /login и статических страниц) защищены.

    📝 Управление задачами
    Создание, чтение, обновление (название/статус) и удаление задач. Каждая задача содержит ID, название, статус выполнения и время создания.

    📊 Статистика
    Получение количества всех, выполненных и активных задач.

    🔎 Фильтрация и сортировка
    Список задач можно фильтровать по статусу (все / активные / выполненные) и сортировать по дате создания или названию (по возрастанию/убыванию).

    🌐 Веб-интерфейс
    Готовый адаптивный интерфейс с матричным фоном, поддерживающий все операции, включая inline-редактирование названия задачи.

    💾 Постоянное хранение
    Данные сохраняются в SQLite (файл todos.db), что обеспечивает сохранность после перезапуска.

    ❤️ Health‑check и метрики
    Эндпоинт /health для мониторинга и /metrics (готов к интеграции с Prometheus).

    🔄 CI/CD
    Автоматическая сборка, тестирование и деплой через GitHub Actions при каждом пуше.

🛠 Стек технологий
Компонент	Технологии
Язык	Go 1.22+
Веб-фреймворк	Стандартный net/http
База данных	SQLite (драйвер mattn/go-sqlite3)
Аутентификация	JWT (golang-jwt/jwt/v5) + bcrypt для хеширования паролей
Фронтенд	Чистый HTML/CSS/JS (без внешних зависимостей)
Тестирование	Встроенный testing + httptest
Контейнеризация	Docker (многоступенчатая сборка)
CI/CD	GitHub Actions
⚡ Быстрый старт
Локальный запуск (без Docker)

    Клонируйте репозиторий:
    bash

    git clone https://github.com/your-username/todo-manager.git
    cd todo-manager

    Установите зависимости:
    bash

    go mod download

    Настройте переменные окружения (опционально):

        PORT — порт для сервера (по умолчанию 8080)

        DB_FILE — путь к файлу SQLite (по умолчанию todos.db)

        JWT_SECRET — секретный ключ для подписи JWT (обязательно измените! по умолчанию используется default-secret-change-me)

    Пример (Linux/macOS):
    bash

    export PORT=8080
    export JWT_SECRET=my-super-secret-key

    Запустите сервер:
    bash

    go run cmd/server/main.go

    Сервер будет доступен на http://localhost:8080.

    Откройте веб-интерфейс
    Перейдите по адресу http://localhost:8080/neiroslop (или /handmade, /neiroslop2 — все ведут на одну и ту же страницу).

Запуск через Docker

Сборка и запуск образа (если Dockerfile присутствует):
bash

docker build -t todo-manager .
docker run -p 8080:8080 -e JWT_SECRET=my-secret todo-manager

📌 API Эндпоинты
Метод	Путь	Описание	Аутентификация
POST	/register	Регистрация нового пользователя	❌
POST	/login	Вход, получение JWT-токена	❌
GET	/health	Проверка работоспособности	❌
GET	/todos	Получение списка задач (с фильтрацией/сортировкой)	✅
POST	/todos	Создание новой задачи	✅
PATCH	/todos/{id}	Обновление задачи (название или статус)	✅
DELETE	/todos/{id}	Удаление задачи по ID	✅
GET	/todos/stats	Получение статистики	✅
GET	/neiroslop	Веб-интерфейс (основная страница)	❌
GET	/handmade	Веб-интерфейс (альтернативный)	❌
GET	/neiroslop2	Веб-интерфейс (альтернативный)	❌

    Примечание: Защищённые эндпоинты требуют передачи JWT-токена в заголовке Authorization: Bearer <token>.

📝 Примеры запросов
Регистрация пользователя
bash

curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"secret123"}'

Ответ:
json

{
  "user_id": 1,
  "username": "john_doe"
}

Вход (получение токена)
bash

curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"secret123"}'

Ответ:
json

{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}

Создание задачи (с токеном)
bash

curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ваш_токен>" \
  -d '{"title":"Купить хлеб"}'

Ответ:
json

{
  "id": 1,
  "title": "Купить хлеб",
  "completed": false,
  "created_at": "2025-03-16T12:34:56Z"
}

Получение списка задач с фильтрацией и сортировкой
bash

curl "http://localhost:8080/todos?filter=active&sort=title&order=asc" \
  -H "Authorization: Bearer <ваш_токен>"

Обновление задачи (статус)
bash

curl -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ваш_токен>" \
  -d '{"completed":true}'

Удаление задачи
bash

curl -X DELETE http://localhost:8080/todos/1 \
  -H "Authorization: Bearer <ваш_токен>"

Получение статистики
bash

curl http://localhost:8080/todos/stats \
  -H "Authorization: Bearer <ваш_токен>"

Ответ:
json

{
  "total": 5,
  "completed": 2,
  "pending": 3
}

📁 Структура проекта
text

todo-manager/
├── cmd/
│   ├── sandbox/               # Статические HTML-страницы (фронтенд)
│   │   ├── index2.html
│   │   ├── index3.html
│   │   └── index4.html
│   └── server/
│       └── main.go            # Точка входа (настройка роутов, сервер)
├── internal/
│   ├── handlers/              # HTTP-обработчики (пока пуст, логика в main.go)
│   └── storage/               # Работа с базой данных
│       ├── storage.go         # Функции CRUD, статистика, пользователи
│       └── models.go          # Структуры данных (DBtodo, UpdateTodoRequest)
├── Dockerfile                 # (планируется)
├── docker-compose.yml         # (планируется)
├── go.mod
├── go.sum
└── .github/
    └── workflows/
        └── ci.yml             # GitHub Actions пайплайн

🧪 Тестирование

Запуск всех тестов:
bash

go test ./...

Проверка покрытия кода:
bash

go test -cover ./...

🔄 CI/CD

При каждом пуше в ветку main или develop GitHub Actions автоматически выполняет:

    Линтинг (golangci-lint)

    Юнит-тесты

    Сборку Docker-образа

    Публикацию образа в GitHub Container Registry (или Docker Hub)

Это гарантирует, что код всегда находится в работоспособном состоянии и готов к деплою.
📦 Docker

Контейнеризация позволяет легко развернуть сервис в любой среде. Dockerfile использует многоступенчатую сборку для минимизации размера образа. При необходимости вы можете настроить переменные окружения через -e или файл .env.

Пример запуска с пользовательскими настройками:
bash

docker run -d \
  -p 8080:8080 \
  -e JWT_SECRET=my-production-secret \
  -v ./data:/app/data \
  todo-manager

Разработано с ❤️ для упрощения жизни разработчиков.
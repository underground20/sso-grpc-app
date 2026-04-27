# gRPC SSO Service

Этот проект представляет собой сервис аутентификации и авторизации (SSO), построенный на Go с использованием gRPC для основного API и HTTP для административных задач. Проект полностью контейнеризирован с помощью Docker и настроен для непрерывной интеграции с GitHub Actions.

## Основные технологии

- **Язык:** Go
- **API:** gRPC, HTTP (для админки)
- **База данных:** PostgreSQL
- **Контейнеризация:** Docker, Docker Compose
- **Тестирование:** Встроенный пакет `testing`, `testify` для ассертов
- **CI/CD:** GitHub Actions
- **Линтинг:** `golangci-lint`

## Структура проекта

- `cmd/sso`: Точка входа в приложение.
- `internal/`: Основная бизнес-логика, разделенная по доменам (auth, admin, storage).
- `migrations/`: SQL-миграции для базы данных.
- `tests/`: Функциональные (интеграционные) тесты.
- `.github/workflows/`: Конфигурация CI/CD для GitHub Actions.
- `docker-compose.yml`: Конфигурация для запуска в режиме разработки.
- `docker-compose.test.yml`: Изолированная конфигурация для запуска тестов.

---

## Запуск для разработки

Для запуска проекта в режиме разработки используется `docker-compose`.

1. **Создайте файл `.env`** в корне проекта. Вы можете скопировать `.env.example` (если он есть) или создать новый:

   ```env
   # .env
   ENV=dev
   GRPC_PORT=44044
   HTTP_PORT=8080

   # Database
   DB_USER=myuser
   DB_PASSWORD=mypassword
   DB_NAME=sso_db
   DATABASE_URL=postgres://myuser:mypassword@postgres:5432/sso_db?sslmode=disable
   ```

2. **Запустите сервисы:**

   ```sh
   docker-compose up --build
   ```

   Эта команда соберет образ вашего Go-приложения, запустит контейнер с PostgreSQL, применит миграции и запустит gRPC-сервер на порту `44044` и HTTP-сервер на порту `8080`.

---

## Тестирование

Проект настроен для запуска функциональных тестов в изолированном окружении, которое также управляется через Docker Compose.

1. **Убедитесь, что у вас есть файл `tests/.env`**. Он необходим для конфигурации тестовой среды. Если его нет, создайте его:

   ```env
   # .env
   ENV=test
   GRPC_PORT=44044
   HTTP_PORT=8080

   # Test Database
   DB_USER=testuser
   DB_PASSWORD=testpassword
   DB_NAME=testdb
   DATABASE_URL=postgres://testuser:testpassword@postgres:5432/testdb?sslmode=disable
   ```

2. **Запустите тестовые сервисы** в фоновом режиме:

   ```sh
   docker-compose -f docker-compose.test.yml up --build -d
   ```

3. **Выполните тесты:**

   ```sh
   go test -v ./...
   ```

4. **Остановите тестовые сервисы** после завершения:

   ```sh
   docker-compose -f docker-compose.test.yml down
   ```

---

## CI/CD

Проект использует GitHub Actions для автоматического линтинга и тестирования. Workflow находится в файле `.github/workflows/ci.yml`.

Процесс CI включает следующие шаги:
1. Запуск линтера `golangci-lint` для статического анализа кода.
2. Сборка и запуск тестового окружения с помощью `docker-compose.test.yml`.
3. Ожидание полной готовности базы данных.
4. Запуск функциональных тестов.
5. Остановка и очистка тестового окружения.

---

## API Endpoints

### gRPC API

Основной API для взаимодействия с сервисом.

- **`auth.Register`**: Регистрирует нового пользователя.
- **`auth.Login`**: Аутентифицирует пользователя и возвращает JWT-токен.
- **`auth.GetRoles`**: Возвращает список всех доступных ролей и их разрешений.

### HTTP Admin API

Внутренний API для административных задач.

- **`POST /admin/app`**: Создает новое приложение-клиент.
  - **Body:**
    ```json
    {
      "name": "my-new-app",
      "secret": "a-very-secret-key"
    }
    ```

- **`POST /admin/role`**: Создает новую роль с набором разрешений.
  - **Body:**
    ```json
    {
      "name": "editor",
      "permissions": ["create-post", "edit-post"]
    }
    ```

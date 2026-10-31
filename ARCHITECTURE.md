# Архитектура и технологии Tracking App

## 1. Технологии разработки

### Backend Framework & Language
- **Go 1.25.0** — основной язык разработки
- **http.ServeMux** — встроенный Go мультиплексор для маршрутизации HTTP запросов

### База данных
- **PostgreSQL 18.1** — основная реляционная БД для хранения данных
- **PGX v5** — высокопроизводительный драйвер PostgreSQL для Go с поддержкой пулинга соединений

### Инструменты разработки и управления

#### Docker & Docker Compose
Используется для изоляции и управления окружением:
- **trackingapp-postgres** — контейнер PostgreSQL для локальной разработки
- **trackingapp-postgres-migrate** — контейнер для выполнения миграций БД
- **port-forwarder** — контейнер alpine/socat для проксирования портов PostgreSQL на localhost

#### Makefile
Команды для управления проектом:
- `env-up` — запуск PostgreSQL контейнера
- `env-down` — остановка контейнера
- `env-cleanup` — полная очистка данных PostgreSQL
- `migrate-create seq=<name>` — создание новой миграции
- `migrate-up` — выполнить вверх все миграции
- `migrate-down` — откатить последнюю миграцию
- `trackingapp-run` — запуск приложения в режиме разработки

#### Database Migrations
- **migrate/migrate v4.19.1** — инструмент для версионирования и управления миграциями БД
- Миграции хранятся в `/migrations/` в формате `.up.sql` и `.down.sql`

### Логирование
- **Zap (go.uber.org/zap)** — структурированное логирование с поддержкой разных уровней (debug, info, warn, error)
- Логи записываются одновременно в консоль и в файлы с временными метками

### Валидация
- **Validator v10 (go-playground/validator)** — декларативная валидация структур с поддержкой кастомных правил
- Используется для валидации HTTP запросов и доменных моделей

### Прочие зависимости
- **UUID (google/uuid)** — генерация уникальных идентификаторов
- **envconfig (kelseyhightower/envconfig)** — загрузка конфигурации из переменных окружения

---

## 2. Архитектура приложения

### Принципы архитектуры

Приложение построено на основе **Чистой архитектуры (Clean Architecture)** с четким разделением ответственности между слоями. Каждый слой изолирован от других и имеет четко определенные зависимости направленные внутрь.

### Структура слоев

```
┌─────────────────────────────────────────────────────────────┐
│                       HTTP Requests                         │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│          TRANSPORT LAYER (HTTP Handlers)                    │
│  - Обработка HTTP запросов                                  │
│  - Декодирование/кодирование JSON                           │
│  - Валидация входных данных                                 │
│  - Преобразование DTO ↔ Domain                              │
│  (internal/features/*/transport/http/)                      │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│          SERVICE LAYER (Business Logic)                     │
│  - Бизнес-логика приложения                                 │
│  - Валидация доменных моделей                               │
│  - Координация работы между компонентами                     │
│  - Обработка ошибок                                         │
│  (internal/features/*/service/)                             │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│        REPOSITORY LAYER (Data Persistence)                  │
│  - Абстракция доступа к данным                              │
│  - SQL запросы                                              │
│  - Преобразование модели БД ↔ Domain                        │
│  - Управление транзакциями                                  │
│  (internal/features/*/repository/postgres/)                 │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│          DATABASE LAYER (PostgreSQL)                        │
│  - Хранение данных                                          │
│  - Выполнение SQL запросов                                  │
│  - Управление соединениями (Connection Pool)                │
└─────────────────────────────────────────────────────────────┘
```

### Компоненты каждого слоя

#### **1. TRANSPORT LAYER** 
`internal/features/users/transport/http/`

**Ответственность:**
- Прием HTTP запросов
- Валидация входных данных (JSON schema, ограничения типов)
- Преобразование внешних DTO в внутренние доменные модели
- Обработка ошибок и формирование HTTP ответов
- Добавление заголовков и кодов состояния

**Компоненты:**
- `create_user.go` — HTTP хендлер для создания пользователя
  - `CreateUserRequest` — DTO входящего запроса
  - `CreateUserResponse` — DTO исходящего ответа
  - `CreateUser(http.ResponseWriter, *http.Request)` — обработчик

**Пример flow:**
```
POST /api/v1/users
{
  "full_name": "John Doe",
  "phone_number": "+1234567890"
}
    ↓
1. Декодирование JSON в CreateUserRequest
2. Валидация полей (required, min/max length, phone format)
3. Преобразование DTO в domain.User (без ID, т.к. он генерируется БД)
4. Вызов service.CreateUser()
5. Преобразование результата в CreateUserResponse
6. HTTP 201 Status + JSON response
```

#### **2. SERVICE LAYER**
`internal/features/users/service/`

**Ответственность:**
- Реализация бизнес-логики
- Валидация доменных правил (domain.User.Validate())
- Координация между репозиториями и другими сервисами
- Обработка ошибок и их трансформация
- Логирование важных операций

**Компоненты:**
- `service.go` — основной сервис
  - `UsersService` — структура с инъекциями зависимостей (repository)
  - `UsersRepository` — интерфейс, скрывающий реализацию хранения
- `create_user.go` — метод для создания пользователя
  - Валидирует domain.User через `user.Validate()`
  - Делегирует сохранение в repository
  - Возвращает готового пользователя или ошибку

**Преимущества использования интерфейса `UsersRepository`:**
- Слабая связанность: Service не знает о PostgreSQL
- Возможность замены реализации (например, на Mock для тестов)
- Возможность добавить новую реализацию (Redis, MongoDB и т.д.) без изменения Service

#### **3. REPOSITORY LAYER**
`internal/features/users/repository/postgres/`

**Ответственность:**
- Абстракция доступа к данным
- Выполнение SQL запросов
- Преобразование между моделями БД (UserModel) и доменными моделями (domain.User)
- Управление соединениями и таймаутами
- Обработка ошибок БД

**Компоненты:**
- `repository.go` — основной репозиторий
  - `UsersRepository` — структура с пулом соединений
  - `pool` — реализация интерфейса Pool (управление соединениями)
- `create_user.go` — метод вставки пользователя
  - Устанавливает таймаут операции
  - Выполняет INSERT запрос с RETURNING для получения сгенерированного ID
  - Сканирует результат в `UserModel`
  - Преобразует `UserModel` в `domain.User`
- `models.go` — структуры для маппинга результатов БД

**SQL запрос для создания:**
```sql
INSERT INTO trackingapp.users (full_name, phone_number)
VALUES ($1, $2)
RETURNING id, version, full_name, phone_number
```

#### **4. DOMAIN LAYER**
`internal/core/domain/`

**Ответственность:**
- Определение моделей бизнес-сущностей
- Инкапсуляция бизнес-логики сущностей
- Валидация инвариантов модели
- Независимость от деталей реализации (БД, транспорт)

**Компоненты:**
- `user.go` — доменная модель пользователя
  - `User` — структура с полями: ID, Version, FullName, PhoneNumber
  - `NewUser()` — конструктор для инициализированного пользователя
  - `NewUserUninitialized()` — конструктор для новых пользователей (без ID)
  - `Validate()` — валидация доменных инвариантов:
    - FullName: минимум 3, максимум 100 символов
    - PhoneNumber: опциональное, формат +XXXX..., 10-15 символов

**Особенность:**
Доменный User не содержит информацию о том, как он создается или хранится. Это чистая бизнес-модель.

#### **5. CORE LAYER** (Поддерживающие компоненты)
`internal/core/`

Общие компоненты, используемые всеми фичами:

**HTTP Server** (`transport/http/server/`)
- `server.go` — HTTP сервер с поддержкой middleware и graceful shutdown
- `route.go` — описание отдельного маршрута (Method, Path, Handler)
- `router.go` — регистрация и управление маршрутами

**HTTP Middleware** (`transport/http/middleware/`)
- Цепочка middleware для предварительной/постобработки запросов
- Логирование, аутентификация, обработа ошибок и т.д.

**Request/Response** (`transport/http/request/`, `response/`)
- `decode.go` — декодирование и валидация JSON запросов
- `handler.go` — обработка ошибок и формирование ответов
- `writer.go` — сериализация ответов в JSON

**Logger** (`logger/`)
- Структурированное логирование через Zap
- Конфигурация логов (уровень, папка)
- Интеграция в контекст запроса (context)

**Database** (`repository/postgres/pool/`)
- `pool.go` — управление пулом соединений PostgreSQL
- `config.go` — конфигурация подключения, таймауты операций

**Errors** (`errors/`)
- Определение кастомных ошибок приложения

---

## 3. Как это работает в синергии (пример: создание пользователя)

### Полный HTTP цикл создания пользователя

```
1. HTTP REQUEST (Transport Layer)
   ├─ POST /api/v1/users
   └─ Body: {"full_name": "John Doe", "phone_number": "+1234567890"}
        ↓
   
2. DECODE & VALIDATE (Transport Layer)
   ├─ JSON → CreateUserRequest
   ├─ Валидирует: required, min/max, phone format
   └─ Ошибка? → HTTP 400 Bad Request
        ↓
   
3. TRANSFORM DTO → DOMAIN (Transport Layer)
   ├─ CreateUserRequest → domain.User
   │  domain.User{
   │    ID: -1 (uninitialized),
   │    Version: -1 (uninitialized),
   │    FullName: "John Doe",
   │    PhoneNumber: "+1234567890"
   │  }
   └─ Передает в Service Layer
        ↓
   
4. SERVICE LAYER - BUSINESS LOGIC
   ├─ service.CreateUser(ctx, user)
   ├─ Вызывает user.Validate()
   │  ├─ Проверяет FullName длину (3-100)
   │  ├─ Проверяет PhoneNumber формат (+, цифры)
   │  └─ Ошибка? → возвращает error
   └─ Передает валидного user в Repository
        ↓
   
5. REPOSITORY LAYER - DATA PERSISTENCE
   ├─ repository.CreateUser(ctx, user)
   ├─ Создает таймаут контекста (5 сек например)
   ├─ Выполняет SQL:
   │  INSERT INTO trackingapp.users (full_name, phone_number)
   │  VALUES ($1, $2)
   │  RETURNING id, version, full_name, phone_number
   ├─ PostgreSQL генерирует:
   │  - id = 1 (SERIAL PRIMARY KEY)
   │  - version = 1 (DEFAULT 1)
   └─ Сканирует результат в UserModel
        ↓
   
6. TRANSFORM USERMODEL → DOMAIN (Repository Layer)
   ├─ UserModel{ID: 1, Version: 1, ...} 
   └─ → domain.User{ID: 1, Version: 1, ...}
        ↓
   
7. TRANSFORM DOMAIN → DTO (Transport Layer)
   ├─ domain.User → CreateUserResponse
   │  {
   │    "id": 1,
   │    "version": 1,
   │    "full_name": "John Doe",
   │    "phone_number": "+1234567890"
   │  }
   └─ Передает в HTTP Response Writer
        ↓
   
8. HTTP RESPONSE
   ├─ Status: 201 Created
   ├─ Content-Type: application/json
   └─ Body: {...response object...}
```

### Прохождение данных между слоями

```
clients/external
      ↓
CreateUserRequest (DTO) ← HTTP JSON payload
      ↓
domain.User ← Transport Layer converts
      ↓
domain.User → Service Layer validates domain
      ↓
domain.User → Repository Layer persists
      ↓
UserModel (DB row) ← PostgreSQL returns
      ↓
domain.User ← Repository converts back
      ↓
CreateUserResponse (DTO) ← Transport Layer converts
      ↓
HTTP JSON response → clients/external
```

### Преимущества такой архитектуры

1. **Слабая связанность**
   - Transport не знает о PostgeSQL
   - Service не знает о HTTP
   - Разные эндпоинты могут использовать один Service
   - Repository можно заменить без изменения остального

2. **Тестируемость**
   - Можно мокировать Repository для тестирования Service
   - Можно мокировать Service для тестирования Transport
   - Бизнес-логика (domain.User.Validate) тестируется отдельно

3. **Масштабируемость**
   - Новые фичи добавляются по готовому паттерну
   - Core компоненты (Server, Logger, Pool) переиспользуются
   - Легко добавлять новые фичи (tasks, statistics и т.д.)

4. **Maintainability**
   - Ясное разделение ответственности
   - Каждый слой отвечает за одно
   - Легко найти где обрабатывается логика

5. **Независимость от фреймворков**
   - Можно заменить http.ServeMux на gin, echo и т.д.
   - Можно заменить PostgreSQL на MongoDB и т.д.
   - Domain модели не зависят от никаких фреймворков

---

## 4. База данных: структура и миграции

### Схема БД

Приложение использует отдельную схему `trackingapp` для изоляции от других приложений.

**Таблица: trackingapp.users**
```sql
CREATE TABLE trackingapp.users (
    id SERIAL PRIMARY KEY,              -- уникальный идентификатор
    version BIGINT DEFAULT 1,           -- версия для оптимистичного лока
    full_name VARCHAR(100) NOT NULL,    -- ФИ пользователя (3-100 символов)
    phone_number VARCHAR(15)            -- опциональный номер телефона
);
```

**Таблица: trackingapp.tasks** (подготовлена, не реализована)
```sql
CREATE TABLE trackingapp.tasks (
    id SERIAL PRIMARY KEY,
    version BIGINT DEFAULT 1,
    title VARCHAR(100) NOT NULL,           -- название задачи
    description VARCHAR(1000),             -- описание
    completed BOOLEAN NOT NULL,            -- статус завершения
    created_at TIMESTAMPTZ NOT NULL,       -- время создания
    completed_at TIMESTAMPTZ,              -- время завершения
    author_user_id INTEGER NOT NULL,       -- ссылка на создателя (FK)
    FOREIGN KEY (author_user_id) REFERENCES trackingapp.users(id),
    CHECK((completed=FALSE AND completed_at IS NULL) OR 
          (completed=TRUE AND completed_at IS NOT NULL AND completed_at >= created_at))
);
```

### Управление миграциями

Все миграции версионируются в `/migrations/`:

- `000001_init.up.sql` — создание схемы и таблиц
- `000001_init.down.sql` — откат (drop schema)

Команды для работы с миграциями:
```bash
make migrate-up              # Выполнить все невыполненные миграции
make migrate-down            # Откатить последнюю миграцию
make migrate-create seq=add_users_index  # Создать новую миграцию
```

---

## 5. Инициализация приложения

### Startup sequence в main.go

```
1. Signal handle (SIGINT, SIGTERM) ← для graceful shutdown
2. Logger initialization ← структурированное логирование
3. PostgreSQL pool initialization ← управление соединениями
4. Features initialization (Users feature)
   ├─ UsersRepository (pool) ← Repository Layer
   ├─ UsersService (repository) ← Service Layer
   └─ UsersHTTPHandler (service) ← Transport Layer
5. HTTP Server start
   ├─ Register routes/handlers
   └─ Listen on configured port (например :8080)
6. Graceful shutdown on signal
   ├─ Close pool
   └─ Close logger
```

### Dependency Injection Pattern

Приложение использует конструкторы для инъекции зависимостей:

```go
// Repository зависит от Pool
usersRepository := users_postgres_repository.NewUsersRepository(pool)

// Service зависит от Repository (через интерфейс)
usersService := users_service.NewUsersService(usersRepository)

// HTTP Handler зависит от Service
usersHTTPHandler := user_transport_http.NewUsersHTTPHandler(usersService)
```

Это позволяет:
- Явно видеть зависимости
- Легко подменять реализации
- Тестировать с моками

---

## 6. Конфигурация

Приложение использует переменные окружения для конфигурации через `envconfig`:

**Database:**
- `POSTGRES_USER` — пользователь БД
- `POSTGRES_PASSWORD` — пароль БД
- `POSTGRES_HOST` — хост (localhost для разработки)
- `POSTGRES_DB` — имя БД

**HTTP Server:**
- `HTTP_ADDR` — адрес и порт (например :8080)

**Logger:**
- `LOGGER_FOLDER` — папка для логов
- `LOGGER_LEVEL` — уровень логирования (debug, info, warn, error)

Загружаются через функции `NewConfig***Must()` которые паникуют при отсутствии обязательных переменных.

---

## Заключение

Приложение построено на принципах Clean Architecture с четким разделением слоев, что обеспечивает:
- ✅ Легкость разработки новых фич по готовому паттерну
- ✅ Простоту тестирования каждого компонента отдельно
- ✅ Возможность замены любого компонента без влияния на остальные
- ✅ Масштабируемость приложения при добавлении новых фич (tasks, statistics и т.д.)

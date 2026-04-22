# Auto HTTP Fetcher

Система управления сервисами и их вебхуками. У каждого из сервисов может быть несколько вебхуков. Данный сервис предоставляет возможность ручного или запланированного запуска вебхуков.

## Архитектура
В проекте используется микросервисная архитектура: 

- **API Gateway** - микросервис, который является единственной точкой входа 
- **Auth** - микросервис авторизации пользователей
- **Module** - микросервис, отвечающий за работу с модулями сервиса 
- **Scheduler** - микросервис-планировщик, отвечающий за порядок выполнения вебхуков
- **Fetcher** - микросервис, отвечающий за выполнение вебхуков и сохранение их результатов
- **Analytics** - микросервис аналитики

Все запросы пользователя попадают в API Gateway, который проксирует их на нужные ресурсы. Взаимодействие между микросервисами происходит с помощью gRPC, а общение между пользователем и приложение происходит по HTTP

## Технологический стек

- **Go 1.23+** - основной язык разработки
- **gRPC** - взаимодействие между микросервисами
- **HTTP/JSON** - общение с клиентами
- **PostgreSQL** - основная база данных
- **Kafka** - брокер сообщений для вебхуков
- **Redis** - кэширование аналитики
- **Docker & Docker Compose** - контейнеризация
- **Goose** - миграции базы данных

## Переменные окружения

### Общие переменные

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `ENV` | Окружение (Development/Production) | `Development` |

### PostgreSQL

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `POSTGRES_URL` | URL подключения к PostgreSQL | `postgres://postgres:postgres@postgres:5432/auto-http-fetcher?sslmode=disable` |

### Redis

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `REDIS_URL` | URL подключения к Redis | `redis:6379` |

### Kafka

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `KAFKA_BROKER` | Адрес Kafka брокера | `kafka:9092` |
| `KAFKA_TOPIC_SCHEDULE_REQUEST` | Топик для запросов планировщика | `schedule-requests` |
| `KAFKA_CONSUMER_GROUP` | Группа потребителей Kafka | `auto-http-fetcher-scheduler` |

### Analytics Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `ANALYTICS_PORT` | Порт для HTTP сервера | `:8080` |
| `ANALYTICS_TTL` | Время жизни кэша в секундах | `60` |

### Fetcher Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `GRPC_PORT` | Порт для gRPC сервера | `:50051` |

### Scheduler Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `FETCHER_GRPC_ADDR` | Адрес gRPC сервера Fetcher | `fetcher:50051` |

### Modules Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `MODULES_PORT` | Порт для HTTP сервера | `:8090` |
| `JWT_SECRET` | Секретный ключ для JWT | `development-secret` |
| `JWT_TTL` | Время жизни JWT токена | `5h` |

### Users Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `USERS_PORT` | Порт для HTTP сервера | `:8095` |
| `JWT_SECRET` | Секретный ключ для JWT | `development-secret` |
| `JWT_TTL` | Время жизни JWT токена | `5h` |

### Gateway Service

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `FETCHER_ADDR` | Адрес Fetcher сервиса | `fetcher:50051` |
| `ANALYTICS_ADDR` | Адрес Analytics сервиса | `analytics:8080` |
| `USERS_ADDR` | Адрес Users сервиса | `users:8095` |
| `MODULES_ADDR` | Адрес Modules сервиса | `modules:8090` |

## Установка и запуск

### 1. Клонирование репозитория
```
git clone https://gitlab.crja72.ru/golang/2026/spring/projects/go12/auto-http-fetcher/
```

### 2. Настройка окружения
Необходимо заполнить файл .env.example и скопировать в .env файл
```
cp .env.example .env
```

Пример .env файла:
```
POSTGRES_URL=postgres://postgres:postgres@localhost:5432/auto-http-fetcher?sslmode=disable
REDIS_URL=redis://localhost:6379
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC_SCHEDULE_REQUEST=schedule-requests
KAFKA_CONSUMER_GROUP=auto-http-fetcher-scheduler
GRPC_PORT=:50051
ANALYTICS_PORT=:8080
MODULES_PORT=:8090
USERS_PORT=:8095
JWT_SECRET=your-secret-key
JWT_TTL=5h
ANALYTICS_TTL=60
ENV=Development
```

## Запуск в Docker
```
docker-compose up -d
```

### Просмотр логов
```
docker-compose logs -f
```
### Логи конкретного сервиса
```
docker-compose logs -f fetcher
docker-compose logs -f analytics
docker-compose logs -f scheduler
docker-compose logs -f gateway
```

### Остановка всех сервисов
```
docker-compose down
```

### Остановка с удалением томов
```
docker-compose down -v
```

### Запуск миграций через docker-compose
```
docker-compose up migrations
```

### Принудительный перезапуск миграций
```
docker-compose up --force-recreate migrations
```

## API Endpoints

### API Gateway (порт 8081)

#### Auth endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/auth/register` | Регистрация нового пользователя |
| POST | `/api/v1/auth/login` | Авторизация пользователя |

#### Modules endpoints (требуют JWT авторизации)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/module/` | Создание модуля |
| PUT | `/api/v1/module/` | Обновление модуля |
| DELETE | `/api/v1/module/{id}` | Удаление модуля по ID |
| GET | `/api/v1/module/{id}` | Получение модуля по ID |
| GET | `/api/v1/modules/` | Получение списка всех модулей |

#### Webhook endpoints (требуют JWT авторизации)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/module/{module_id}/webhook/` | Создание вебхука для модуля |
| PUT | `/api/v1/module/{module_id}/webhook/` | Обновление вебхука модуля |
| DELETE | `/api/v1/module/{module_id}/webhook/{webhook_id}` | Удаление вебхука |

#### Analytics endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/analytics` | Получение статистики по вебхукам |

#### Fetcher endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/fetcher/fetch` | Выполнение HTTP запроса |
| GET | `/api/v1/fetcher/status` | Получение статуса фетчера |

## Тестирование

### Запуск тестов
```
go test -v ./...
```

### Запуск тестов с покрытием
```
go test -cover ./...
```
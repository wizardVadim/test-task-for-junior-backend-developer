# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8089`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` и `migrations/0002_create_task_recurrences.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8089/swagger/
```

OpenAPI JSON:

```text
http://localhost:8089/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`


## Versions

### 0.0.1

Начальная версия сервиса трекера задач.

Реализован базовый CRUD-функционал для работы с задачами:

- создание задачи (`POST /tasks`)
- получение списка задач (`GET /tasks`)
- получение задачи по ID (`GET /tasks/{id}`)
- обновление задачи (`PUT /tasks/{id}`)
- удаление задачи (`DELETE /tasks/{id}`)

### Модель задачи

Задача содержит следующие поля:

- `id` — уникальный идентификатор
- `title` — название задачи
- `description` — описание задачи
- `status` — текущий статус задачи (`new`, `in_progress`, `done`)
- `created_at` — дата создания
- `updated_at` — дата последнего обновления

### Архитектура

Проект реализован с разделением на слои:

- `domain` — доменные модели
- `usecase` — бизнес-логика
- `repository` — работа с базой данных (PostgreSQL)
- `transport/http` — HTTP API (handlers, router, DTO)
- `infrastructure` — подключение к базе

### Инфраструктура

- PostgreSQL в качестве базы данных
- Docker Compose для запуска окружения
- Swagger/OpenAPI для документации API
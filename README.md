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

Причина в том, что SQL-файлы из директории `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

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

### Задачи

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

### Экземпляры задач (occurrences)

- `GET /api/v1/occurrences?date=YYYY-MM-DD`
- `GET /api/v1/tasks/{id}/occurrences`
- `PATCH /api/v1/occurrences/{id}`

## Поддерживаемые типы периодичности

- `daily`
- `monthly`
- `specific_dates`
- `even_days`
- `odd_days`

## Генерация экземпляров задач

- генерация каждые 5 минут для новых/изменённых задач
- ежедневная генерация (01:00 UTC)
- защита от дублей через `(task_id, scheduled_date)`

## Versions

### 0.0.1

Базовый CRUD для задач.

### 0.0.2

Добавлено:

- периодичность задач
- типы: daily, monthly, specific_dates, even_days, odd_days
- таблицы: task_recurrences, task_recurrence_dates
- генерация экземпляров задач (cron через goroutines)
- сущность task_occurrence
- API для occurrences
- обновление статусов экземпляров
- синхронизация статуса задачи
- улучшенный Swagger

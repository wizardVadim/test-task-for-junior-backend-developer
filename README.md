# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose


## Настройки генерации задач

Конфигурация задаётся через переменные окружения (в docker-compose.yml):

- `GENERATOR_RECENT_INTERVAL_MINUTES` — как часто запускается генератор новых задач (в минутах)
- `GENERATOR_RECENT_WINDOW_MINUTES` — окно времени для поиска недавно созданных/обновлённых задач
- `GENERATOR_RECENT_LOOKAHEAD_DAYS` — на сколько дней вперёд генерируются задачи в частом генераторе
- `GENERATOR_DAILY_LOOKAHEAD_DAYS` — на сколько дней вперёд генерируются задачи в ежедневном генераторе
- `GENERATOR_DAILY_RUN_HOUR` — час запуска ежедневного генератора (UTC)
- `GENERATOR_DAILY_RUN_MINUTE` — минута запуска ежедневного генератора (UTC)

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8089`.

1. Запустить тест создания задач:

```bash
k6 run load-tests/create_tasks.js
```

2. Запустить тест получения occurrences:

```bash
k6 run load-tests/occurrences.js
```

## Перезапуск

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

## Архитектурные решения

- Задача (`task`) рассматривается как шаблон
- Для выполнения используются экземпляры задач (`task_occurrence`)
- Периодичность хранится отдельно (`task_recurrences`)
- Конкретные даты — в `task_recurrence_dates`

Экземпляры задач не создаются сразу на все будущие даты, а генерируются:
- либо при создании/обновлении задачи (частый генератор)
- либо ежедневно (фоновой генератор)

Это позволяет:
- избежать избыточного хранения данных
- снизить нагрузку на базу
- гибко реагировать на изменения правил

### Особенности

- Задачи без периодичности также получают экземпляр (`task_occurrence`) на текущую дату

## Генерация экземпляров задач

- частый генератор:
  - запускается каждые N минут
  - обрабатывает недавно созданные/обновлённые задачи
  - генерирует задачи на ближайшие дни

- ежедневный генератор:
  - запускается по расписанию (по умолчанию 01:00 UTC)
  - генерирует задачи на несколько дней вперёд

- защита от дублей:
  - уникальный индекс `(task_id, scheduled_date)`

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

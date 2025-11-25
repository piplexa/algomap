# ТЗ: Backend API

## 1. Общее описание
REST API для управления схемами и их выполнением.

## 2. Технологический стек
- Golang 1.21+
- HTTP router (chi / gin / fiber - TBD)
- PostgreSQL driver (pgx)
- JWT для аутентификации
- OpenAPI/Swagger документация

## 3. Аутентификация
TODO: определить механизм (JWT, session, API keys?)

## 4. Endpoints

### 4.1 Схемы
```
GET    /api/schemas          - список схем
GET    /api/schemas/:id      - получить схему
POST   /api/schemas          - создать схему
PUT    /api/schemas/:id      - обновить схему
DELETE /api/schemas/:id      - удалить схему
```

### 4.2 Выполнение
```
POST   /api/executions                    - запустить схему (manual)
GET    /api/executions/:id                - статус выполнения
POST   /api/executions/:id/pause          - пауза
POST   /api/executions/:id/resume         - продолжить
POST   /api/executions/:id/stop           - остановить
GET    /api/executions/:id/steps          - история шагов
GET    /api/executions/:id/logs           - логи выполнения
```

### 4.3 Webhook
```
POST   /webhook/:schema_id                - запуск через webhook
POST   /webhook/:schema_id/:step_id       - запуск с конкретного шага
```

### 4.4 Мета-информация
```
GET    /api/node-types                    - список доступных типов нод
GET    /api/node-types/:type              - описание конкретного типа
```

## 5. Формат данных

### 5.1 Schema JSON
```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "status": "draft|active|archived",
  "nodes": [...],
  "edges": [...],
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### 5.2 Execution Response
```json
{
  "id": "uuid",
  "schema_id": "uuid",
  "status": "running|paused|completed|failed",
  "current_step": "node_id",
  "started_at": "timestamp",
  "finished_at": "timestamp|null",
  "error": "string|null"
}
```

## 6. Обработка ошибок
- Стандартные HTTP коды
- JSON формат ошибок
- Валидация входных данных

## 7. Rate limiting
TODO: нужен ли для MVP?

## 8. Открытые вопросы
- Pagination для списков?
- Websockets для real-time статуса?
- Batch операции?
# Worker Testing Checklist

## Подготовка окружения

- [ ] PostgreSQL запущен и доступен
- [ ] RabbitMQ запущен и доступен
- [ ] База данных `algomap` создана
- [ ] Миграции применены
- [ ] Очередь `schema_execution_queue` создана в RabbitMQ

## Сборка и запуск

- [ ] `make build-worker` успешно собирает бинарник
- [ ] `make run-worker` запускает worker без ошибок
- [ ] В логах видно "worker started, waiting for messages..."
- [ ] Worker подключается к БД (лог "connected to database")

## Базовое выполнение

### Тест 1: Start → End

```sql
-- Создать простую схему
INSERT INTO main.schemas (name, definition, created_by, id_status)
VALUES ('Test Start-End', '{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [
    {"source": "start_1", "target": "end_1"}
  ]
}', 1, 1)
RETURNING id;

-- Создать execution
INSERT INTO main.executions (id, schema_id, id_status, created_by, created_at)
VALUES ('11111111-1111-1111-1111-111111111111', <schema_id>, 1, 1, NOW());
```

```bash
# Опубликовать сообщение в RabbitMQ
docker exec -it <rabbitmq_container> rabbitmqadmin publish \
  routing_key=schema_execution_queue \
  payload='{"execution_id":"11111111-1111-1111-1111-111111111111","schema_id":<schema_id>,"current_node_id":"start_1","debug_mode":false}'
```

Проверки:
- [ ] Worker получил сообщение
- [ ] Выполнил start ноду (лог "executing node")
- [ ] Сохранил execution_step для start
- [ ] Опубликовал сообщение для end_1
- [ ] Выполнил end ноду
- [ ] Сохранил execution_step для end
- [ ] Обновил execution.id_status = 4 (completed)
- [ ] Установил execution.finished_at

```sql
-- Проверить результаты
SELECT * FROM main.executions WHERE id = '11111111-1111-1111-1111-111111111111';
SELECT * FROM main.execution_steps WHERE execution_id = '11111111-1111-1111-1111-111111111111';
SELECT * FROM main.execution_state WHERE execution_id = '11111111-1111-1111-1111-111111111111';
```

### Тест 2: Start → Log → End

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "log_1", "type": "log", "data": {"message": "Hello from worker!", "level": "info"}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [
    {"source": "start_1", "target": "log_1"},
    {"source": "log_1", "target": "end_1"}
  ]
}
```

Проверки:
- [ ] Start выполнилась
- [ ] Log выполнилась и записала в логи "Hello from worker!"
- [ ] End выполнилась
- [ ] Все 3 шага сохранены в execution_steps
- [ ] execution завершён успешно

### Тест 3: Start → Variable Set → Log → End

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "var_1", "type": "variable_set", "data": {"variable_name": "test_var", "variable_value": 42}},
    {"id": "log_1", "type": "log", "data": {"message": "Variable set", "level": "info"}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [
    {"source": "start_1", "target": "var_1"},
    {"source": "var_1", "target": "log_1"},
    {"source": "log_1", "target": "end_1"}
  ]
}
```

Проверки:
- [ ] Variable Set выполнилась
- [ ] В execution_state.context появилась переменная test_var = 42
- [ ] Все шаги выполнены
- [ ] execution завершён

### Тест 4: Start → Math → End

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "math_1", "type": "math", "data": {"operation": "add", "left": 10, "right": 5}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [
    {"source": "start_1", "target": "math_1"},
    {"source": "math_1", "target": "end_1"}
  ]
}
```

Проверки:
- [ ] Math выполнилась
- [ ] В output.result = 15
- [ ] execution завершён

## Обработка ошибок

### Тест 5: Несуществующий тип ноды

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "unknown_1", "type": "unknown_type", "data": {}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [...]
}
```

Проверки:
- [ ] Worker логирует ошибку "handler not found for node type"
- [ ] execution.id_status = 5 (failed)
- [ ] execution.error содержит описание

### Тест 6: Невалидная конфигурация ноды

```json
{
  "nodes": [
    {"id": "math_1", "type": "math", "data": {"operation": "divide", "left": 10, "right": 0}}
  ]
}
```

Проверки:
- [ ] Нода возвращает StatusFailed
- [ ] Error = "division by zero"
- [ ] execution завершён с ошибкой

## Sleep нода

### Тест 7: Start → Sleep → End

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "sleep_1", "type": "sleep", "data": {"duration": 5, "unit": "seconds"}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [...]
}
```

Проверки:
- [ ] Sleep вернула StatusSleep
- [ ] SleepUntil установлен на +5 секунд
- [ ] execution.id_status = 3 (paused)
- [ ] TODO: Интеграция с `at` - задача создана
- [ ] TODO: Через 5 секунд execution продолжилось

## Context и переменные

### Тест 8: Проверка context между нодами

```json
{
  "nodes": [
    {"id": "start_1", "type": "start", "data": {}},
    {"id": "var_1", "type": "variable_set", "data": {"variable_name": "x", "variable_value": 10}},
    {"id": "var_2", "type": "variable_set", "data": {"variable_name": "y", "variable_value": 20}},
    {"id": "math_1", "type": "math", "data": {"operation": "add", "left": "{{variables.x}}", "right": "{{variables.y}}"}},
    {"id": "end_1", "type": "end", "data": {}}
  ],
  "edges": [...]
}
```

Проверки:
- [ ] TODO: Интерполяция переменных работает
- [ ] Math получает x=10, y=20
- [ ] Результат = 30

## Performance

### Тест 9: Множественные execution одновременно

Запустить 10 execution параллельно

Проверки:
- [ ] Все execution выполнились
- [ ] Нет гонок в БД
- [ ] Транзакции отработали корректно

### Тест 10: Длинная цепочка нод

Схема с 20+ нодами

Проверки:
- [ ] Все ноды выполнились
- [ ] 20+ записей в execution_steps
- [ ] execution завершён

## Graceful Shutdown

### Тест 11: SIGTERM во время выполнения

1. Запустить worker
2. Опубликовать сообщение с long-running нодой
3. Послать SIGTERM

Проверки:
- [ ] Worker логирует "received shutdown signal"
- [ ] Текущая нода завершается
- [ ] Worker останавливается корректно
- [ ] Состояние сохранено в БД

## TODO: Интеграция с at

- [ ] Sleep нода передаёт задачу в `at`
- [ ] Задача сохраняется в at.db
- [ ] По таймеру `at` вызывает callback
- [ ] Callback публикует сообщение обратно в RabbitMQ
- [ ] Execution продолжается с сохранённого состояния

## Метрики (будущее)

- [ ] Количество обработанных сообщений
- [ ] Время выполнения нод
- [ ] Количество ошибок
- [ ] Размер очереди

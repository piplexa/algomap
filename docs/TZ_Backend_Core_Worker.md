# Выжимка: Архитектура Worker и алгоритм выполнения схем

## Концепция выполнения

**Паттерн: "One node = one message"**
- Каждая нода = одно сообщение в RabbitMQ
- После выполнения ноды воркер публикует новое сообщение со следующей нодой
- Это обеспечивает отсутствие гонок и простую балансировку нагрузки

---

## Алгоритм работы Worker

```
┌─────────────────────────────────────────────────────────────┐
│  Worker получает сообщение из RabbitMQ:                     │
│  {                                                          │
│    "execution_id": "f237bcb0-...",                          │
│    "current_node_id": "start_1",                            │
│    "debug_mode": false                                      │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘
                    ↓
┌───────────────────────────────────────────────────────────────┐
│  1. Загрузить execution_state из БД:                          │
│     SELECT current_node_id, context                           │
│     FROM main.execution_state                                 │
│     WHERE execution_id = 'f237bcb0-...'                       │
│                                                               │
│  Если записи НЕТ → первый запуск → создать начальный context  │
└───────────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────────┐
│  2. Загрузить схему из БД:                                  │
│     SELECT definition FROM main.schemas WHERE id = ...      │
└─────────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────────────┐
│  3. Найти ноду в definition по node_id                          │
│     node = schema.definition.nodes.find(n => n.id == "start_1") │
└─────────────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────────┐
│  4. Выполнить ноду через handler:                           │
│     result = executeNode(node, context)                     │
│                                                             │
│     Возвращает:                                             │
│     {                                                       │
│       "output": {...},                                      │
│       "next_node_id": "http_1",                             │
│       "status": "success/failed/sleep"                      │
│     }                                                       │
└─────────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────────┐
│  5. Сохранить шаг в execution_steps (в транзакции):         │
│     INSERT INTO main.execution_steps (                      │
│       execution_id, node_id, node_type,                     │
│       input, output, id_status,                             │
│       started_at, finished_at                               │
│     ) VALUES (...)                                          │
└─────────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────────┐
│  6. Обновить context:                                       │
│     context.steps[node_id] = {                              │
│       "output": result.output                               │
│     }                                                       │
└─────────────────────────────────────────────────────────────┘
                    ↓
┌───────────────────────────────────────────────────────────────┐
│  7. Сохранить обновлённый state (в той же транзакции):        │
│     INSERT INTO main.execution_state (...)                    │
│     ON CONFLICT (execution_id) DO UPDATE ...                  │
|  TODO: Подумать, получается контекст един для всех шагов, нет │ 
|         истории изменения контекста, п значит получается не   │
|         очень удобно будет делать debug схем                  │
└───────────────────────────────────────────────────────────────┘
                    ↓
┌────────────────────────────────────────────────────────────────┐
│  8. Обработать результат:                                      │
│                                                                │
│  • Если End нода:                                              │
│    - UPDATE executions SET status=completed, finished_at=NOW() │
│    - ACK сообщение                                             │
│    - BREAK                                                     │
│                                                                │
│  • Если Sleep нода:                                            │
│    - Сохранить состояние                                       │
│    - Передать в `at` library для отложенного выполнения        │
│    - ACK сообщение                                             │
│    - BREAK                                                     │
│                                                                │
│  • Если Error:                                                 │
│    - UPDATE executions SET status=failed, error=...            │
│    - ACK сообщение                                             │
│    - BREAK                                                     │
│                                                                │
│  • Если Success:                                               │
│    - Найти next_node через edges                               │
│    - Опубликовать новое сообщение в RabbitMQ:                  │
│      {"execution_id": "...", "current_node_id": "http_1"}      │
│    - ACK текущее сообщение                                     │
│    - BREAK                                                     │
└────────────────────────────────────────────────────────────────┘
```

---

## Структура Context

Context хранится в `execution_state.context` (JSONB) и содержит:

```json
{
  "webhook": {
    "payload": {...}  // Данные от webhook триггера
  },
  "user": {
    "id": 2,
    "email": "user@example.com"
  },
  "execution": {
    "id": "f237bcb0-0b59-48e5-a2a2-85e668748c8b"
  },
  "steps": {
    "start_1": {
      "output": {}
    },
    "http_1": {
      "output": {
        "status_code": 200,
        "body": {"userId": 123, "balance": 500}
      }
    }
  },
  "variables": {
    "user_age": 25,
    "final_price": 100
  }
}
```

**Доступ к переменным:**
- `{{webhook.payload.test}}` - данные от webhook
- `{{user.email}}` - email пользователя
- `{{steps.http_1.output.body.balance}}` - результат HTTP Request
- `{{variables.user_age}}` - переменная установленная Variable Set нодой

---

## Ключевые принципы

### 1. Атомарность шагов
- Каждый шаг выполняется в **транзакции с timeout** (например, 30 сек)
- Сохраняется: `execution_step` + `execution_state` + `executions.current_step_id`
- Если транзакция не успела → откат → можно повторить

### 2. Timeout на ноды
- Если нода работает > N секунд (например, 30) → считается ошибкой
- Context с таймаутом передаётся в handler каждой ноды
- Долгие операции → пользователь должен переделать схему (например, через Sleep)

### 3. Retry логика
- Реализуется внутри каждой ноды (например, HTTP Request имеет свой retry)
- Worker просто выполняет ноду и получает финальный результат

### 4. Поиск следующей ноды
Через edges в schema.definition:

```json
{
  "nodes": [...],
  "edges": [
    {"source": "start_1", "target": "http_1"},
    {"source": "http_1", "target": "condition_1"}
  ]
}
```

После выполнения `http_1` → ищем edge где `source == "http_1"` → берём `target`.

**Для Condition ноды:**
- Возвращает несколько возможных next_node_id
- Worker выбирает нужный на основе результата condition

---

## Таблицы БД

### execution_state (Контекст выполнения)
```sql
CREATE TABLE main.execution_state (
    execution_id UUID PRIMARY KEY,
    current_node_id VARCHAR(255) NOT NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### execution_steps (ВЫполненные шаги схемы)
```sql
CREATE TABLE main.execution_steps (
    id BIGSERIAL PRIMARY KEY,
    execution_id UUID NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    prev_node_id VARCHAR(255),
    next_node_id VARCHAR(255),
    input JSONB,
    output JSONB,
    id_status SMALLINT NOT NULL,  -- 1=success, 2=failed, 3=skipped
    error TEXT,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP
);
```

### executions (Основная таблица выполнения схем)
```sql
CREATE TABLE main.executions (
    id UUID PRIMARY KEY,
    schema_id BIGINT NOT NULL,
    id_status SMALLINT NOT NULL,  -- 1=pending, 2=running, 3=paused, 4=completed, 5=failed, 6=stopped
    current_step_id VARCHAR(255),  -- для UI (показать прогресс)
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    created_by BIGINT NOT NULL,
    error TEXT
);
```

---

## RabbitMQ очередь

**Название:** `schema_execution_queue`

**Формат сообщения:**
```json
{
  "execution_id": "uuid",
  "schema_id": 123,
  "current_node_id": "start",
  "debug_mode": false
}
```

**Настройки:**
- `durable: true` - очередь переживёт перезапуск RabbitMQ
- `delivery_mode: persistent` - сообщения на диск
- `ack: manual` - воркер подтверждает обработку

---

## Плюсы архитектуры "one node = one message"

1. ✅ **Нет гонки** - один execution всегда обрабатывает только один воркер
2. ✅ **Простая балансировка** - добавил воркеров → больше throughput
3. ✅ **Надёжность** - если воркер упал → NACK → другой подхватит
4. ✅ **Прозрачность** - видно сколько схем в работе
5. ✅ **Sleep естественно** - просто не кидаем сразу обратно, а через `at`
6. ✅ **Debug режим** - легко реализовать step-by-step выполнение

---

## Типы нод для реализации

```go
const (
	NodeTypeStart          = "start"           // Стартовая точка
	NodeTypeEnd            = "end"             // Завершение
	NodeTypeLog            = "log"             // Логирование
	NodeTypeHTTPRequest    = "http_request"    // HTTP запрос
	NodeTypeCondition      = "condition"       // Условие (if/else)
	NodeTypeVariableSet    = "variable_set"    // Установка переменной
	NodeTypeMath           = "math"            // Математика
	NodeTypeSleep          = "sleep"           // Задержка
	NodeTypeRabbitMQPublish = "rabbitmq_publish" // Публикация в очередь
)
```

Каждая нода реализует интерфейс:
```go
type NodeHandler interface {
    Execute(ctx context.Context, node *Node, context map[string]interface{}) (*NodeResult, error)
}

type NodeResult struct {
    Output     map[string]interface{}
    NextNodeID *string
    Status     string  // "success", "failed", "sleep"
    Error      *string
    SleepUntil *time.Time  // для Sleep ноды
}
```

---

## Debug режим

Когда `debug_mode: true`:
1. Worker выполняет ноду
2. Сохраняет результат в БД
3. **НЕ публикует** следующее сообщение
4. Меняет `executions.id_status = waiting_debug`
5. Frontend показывает результаты и кнопку "Continue"
6. При нажатии "Continue" → API публикует следующее сообщение

---

## Sleep нода и `at` library

Когда нода возвращает `Status: "sleep"`:
1. Сохраняем state в БД
2. Передаём в `at` library: `at.Schedule(sleep_until, callback)`
3. ACK сообщение из RabbitMQ
4. Когда время наступит → `at` вызывает callback → публикуем обратно в RabbitMQ

**Почему `at` а не RabbitMQ delayed plugin?**
- Надёжность - `at` хранит в Postgresql и самое главное это вообще сторонний сервис (условно надежный), переживёт перезапуск
- RabbitMQ delayed plugin имеет проблемы при больших задержках

---

## Что нужно реализовать в Worker

1. **Consumer** - получение сообщений из RabbitMQ
2. **Engine** - основной движок выполнения (алгоритм выше)
3. **Node Handlers** - обработчики для каждого типа нод
4. **Context Manager** - работа с переменными и interpolation `{{...}}`
5. **Edge Resolver** - поиск следующей ноды через edges
6. **Transaction Manager** - атомарное сохранение шагов
7. **At Integration** - интеграция с сервисом `at` для Sleep

---

## Пример последовательности выполнения

```
Схема: Start → HTTP Request → Condition → Variable Set → End

1. Сообщение: {execution_id: "...", current_node_id: "start"}
   → Выполняет Start
   → Публикует: {execution_id: "...", current_node_id: "http_1"}

2. Сообщение: {execution_id: "...", current_node_id: "http_1"}
   → Выполняет HTTP Request
   → context.steps.http_1.output = {status_code: 200, ...}
   → Публикует: {execution_id: "...", current_node_id: "condition_1"}

3. Сообщение: {execution_id: "...", current_node_id: "condition_1"}
   → Проверяет условие: {{steps.http_1.output.status_code}} == 200
   → true → next_node = "var_set_1"
   → Публикует: {execution_id: "...", current_node_id: "var_set_1"}

4. Сообщение: {execution_id: "...", current_node_id: "var_set_1"}
   → Устанавливает переменную
   → context.variables.result = "success"
   → Публикует: {execution_id: "...", current_node_id: "end_1"}

5. Сообщение: {execution_id: "...", current_node_id: "end_1"}
   → Выполняет End
   → UPDATE executions SET status=completed
   → ACK, завершение
```

---

## Структура проекта Worker

```
backend/
├── cmd/
│   └── worker/
│       └── main.go          # Точка входа worker
│
├── internal/
│   ├── executor/            # Движок выполнения
│   │   ├── engine.go        # Основной алгоритм
│   │   ├── context.go       # Работа с context
│   │   └── edges.go         # Поиск следующей ноды
│   │
│   └── nodes/               # Обработчики нод
│       ├── handler.go       # Интерфейс NodeHandler
│       ├── start.go
│       ├── end.go
│       ├── http.go
│       ├── condition.go
│       ├── variable_set.go
│       ├── math.go
│       ├── sleep.go
│       └── log.go
│
└── pkg/
    └── rabbitmq/
        └── consumer.go      # Consumer для RabbitMQ
```
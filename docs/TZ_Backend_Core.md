# ТЗ: Backend Core (Execution Engine)

## 1. Общее описание
Движок выполнения схем - сердце системы. Обрабатывает задачи из RabbitMQ, выполняет ноды, сохраняет состояние.

## 2. Технологический стек
- Golang 1.21+
- PostgreSQL (pgx)
- RabbitMQ (amqp091-go)
- github.com/piplexa/at (отложенные задачи)
- SQLite (для персистентности `at`)

## 3. Архитектура выполнения

### 3.1 Основной цикл
```
1. Получить задачу из RabbitMQ
2. Загрузить схему из БД
3. Загрузить состояние выполнения
4. Найти текущую ноду
5. Выполнить ноду
6. Сохранить результат в БД (атомарно!)
7. Определить следующую ноду
8. Если есть следующая нода → goto 4
9. Если конец → завершить выполнение
10. ACK сообщение в RabbitMQ
```

### 3.2 Обработка Sleep/Wait
```
1. Нода возвращает "sleep" + duration
2. Сохранить состояние в БД
3. Передать задачу в библиотеку `at`
4. ACK сообщение в RabbitMQ
5. По истечении таймера `at` отправляет задачу обратно в RabbitMQ
6. Выполнение продолжается с сохранённого шага
```

### 3.3 Обработка ошибок
- Retry логика (настраиваемая на уровне ноды)
- Dead Letter Queue для фатальных ошибок
- Логирование всех шагов

## 4. Структура данных

### 4.1 Execution State (БД)
```sql
CREATE TABLE execution_state (
    execution_id UUID PRIMARY KEY,
    current_node_id VARCHAR(255),
    context JSONB,  -- переменные, доступные в схеме
    status VARCHAR(50),
    updated_at TIMESTAMP
);
```

### 4.2 Execution Steps (история)
```sql
CREATE TABLE execution_steps (
    id UUID PRIMARY KEY,
    execution_id UUID,
    node_id VARCHAR(255),
    node_type VARCHAR(50),
    input JSONB,
    output JSONB,
    status VARCHAR(50),  -- success|failed|skipped
    error TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP
);
```

### 4.3 RabbitMQ Message Format
```json
{
  "execution_id": "uuid",
  "schema_id": "uuid",
  "action": "start|resume",
  "context": {...}  // начальные переменные
}
```

## 5. Типы нод (обработчики)

### 5.1 Базовый интерфейс Node Handler
```go
type NodeHandler interface {
    Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error)
}

type NodeResult struct {
    Output    map[string]interface{}
    NextNodes []string  // куда переходить дальше
    Sleep     *time.Duration  // если нужна задержка
    Error     error
}
```

### 5.2 Список обработчиков
TODO: детали после проработки типов нод

## 6. Контекст выполнения

### 6.1 Доступные переменные внутри схемы
- `webhook.payload` - данные от webhook
- `user.email` - кто запустил схему
- `execution.id` - ID текущего выполнения
- `steps.<node_id>.output` - результаты предыдущих шагов
- `env.*` - системные переменные (опционально)

### 6.2 Интерполяция переменных
```
"url": "https://api.example.com/users/{{user.email}}"
```

## 7. Параллельность и масштабирование
- Несколько воркеров читают из одной очереди RabbitMQ
- Каждое выполнение независимо
- Конкурентный доступ к БД (оптимистичные блокировки?)

## 8. Мониторинг и метрики
- Длительность выполнения нод
- Количество ошибок
- Размер очереди RabbitMQ
- Статус воркеров

## 9. Graceful Shutdown
- Завершить текущие задачи
- NACK незавершённые сообщения
- Закрыть соединения с БД/RabbitMQ

## 10. Открытые вопросы
- Таймауты на выполнение ноды?
- Максимальная длительность одного выполнения?
- Как обрабатывать циклы в схеме?
- Защита от бесконечных циклов?
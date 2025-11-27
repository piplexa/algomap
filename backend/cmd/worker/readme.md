# AlgoMap Worker

Worker для выполнения схем. Получает сообщения из RabbitMQ и выполняет ноды последовательно.

## Архитектура

```
RabbitMQ → Consumer → Engine → Node Handler → БД
                         ↓
                    Publisher → RabbitMQ (следующая нода)
```

### Компоненты

1. **Consumer** (`pkg/rabbitmq/consumer.go`) - получает сообщения из очереди
2. **Engine** (`internal/executor/engine.go`) - движок выполнения
3. **Node Handlers** (`internal/nodes/*.go`) - обработчики типов нод
4. **Publisher** (`pkg/rabbitmq/consumer.go`) - публикует следующие ноды

## Запуск

### Локально

```bash
cd backend
go run cmd/worker/main.go
```

## Конфигурация

Worker читает переменные окружения:

- `DATABASE_URL` - подключение к PostgreSQL
- `RABBITMQ_URL` - подключение к RabbitMQ
- `LOG_LEVEL` - уровень логирования (debug/info/warn/error)

## Алгоритм работы

```
1. Получить сообщение из RabbitMQ:
   {execution_id, schema_id, current_node_id, debug_mode}

2. Загрузить execution_state из БД
   (если нет - создать начальный)

3. Загрузить схему из БД

4. Найти ноду по ID

5. Получить обработчик для типа ноды

6. Выполнить ноду → получить result

7. Сохранить execution_step в БД (в транзакции)

8. Обновить context

9. Найти следующую ноду через edges

10. Сохранить execution_state

11. Обновить статус execution

12. Обработать результат:
    - End нода → завершить, ACK
    - Sleep нода → передать в 'at', ACK
    - Error → завершить с ошибкой, ACK
    - Success → опубликовать следующую ноду, ACK
```

## Типы нод

Сейчас реализованы:

- ✅ `start` - стартовая точка
- ✅ `end` - завершение
- ✅ `log` - логирование

TODO:

- ⏳ `http_request` - HTTP запросы
- ⏳ `condition` - условия (if/else)
- ⏳ `variable_set` - установка переменных
- ⏳ `math` - математические операции
- ⏳ `sleep` - задержка выполнения
- ⏳ `rabbitmq_publish` - публикация в RabbitMQ

## Добавление нового типа ноды

1. Создать файл `internal/nodes/my_node.go`:

```go
package nodes

type MyNodeHandler struct{}

func NewMyNodeHandler() *MyNodeHandler {
    return &MyNodeHandler{}
}

func (h *MyNodeHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
    // Реализация
    return &NodeResult{
        Output: map[string]interface{}{
            "result": "ok",
        },
        Status: StatusSuccess,
    }, nil
}
```

2. Зарегистрировать в `cmd/worker/main.go`:

```go
registry.Register("my_node", nodes.NewMyNodeHandler())
```

## Debug

Worker логирует все действия:

```
{"level":"info","msg":"worker started, waiting for messages..."}
{"level":"info","msg":"executing node","execution_id":"...","node_id":"start_1"}
{"level":"info","msg":"node executed successfully","execution_id":"...","status":"success"}
```

## Graceful Shutdown

Worker корректно завершается по сигналу `SIGTERM` / `SIGINT`:

```
^C
{"level":"info","msg":"received shutdown signal"}
{"level":"info","msg":"worker shutting down..."}
{"level":"info","msg":"worker stopped gracefully"}
```

## TODO

- [ ] Реализовать остальные типы нод
- [ ] Добавить retry логику для failed нод
- [ ] Интеграция с `at` library для Sleep
- [ ] Metrics (Prometheus)
- [ ] Health check endpoint
- [ ] Поддержка debug режима (step-by-step)

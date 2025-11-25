# ТЗ: Типы нод (Node Types)

## 1. Общее описание
Типы нод - это "кирпичики" из которых строятся схемы. Каждый тип решает конкретную задачу.

## 2. Базовые принципы

### 2.1 Каждая нода имеет:
- **ID** - уникальный в рамках схемы
- **Type** - тип ноды
- **Config** - конфигурация (параметры)
- **Inputs** - входящие соединения
- **Outputs** - исходящие соединения

### 2.2 Выходы (outputs) могут быть:
- **Single** - одно направление (обычная нода)
- **Multiple** - несколько направлений (условие, switch)
- **None** - конец выполнения (End нода)

## 3. Категории нод

### 3.1 Управление потоком (Flow Control)
- Start
- End
- Condition (If/Else)
- Switch (multiple branches)
- Loop (циклы)

### 3.2 Данные (Data Operations)
- Variable Set
- Variable Get
- Math Operations
- String Operations
- JSON Operations

### 3.3 Внешние вызовы (External)
- HTTP Request
- RabbitMQ Publish
- Database Query
- (SubSchema - заглушка)

### 3.4 Время (Time)
- Sleep/Wait
- Delay

### 3.5 Логика (Logic)
- Log/Print
- Error/Throw

## 4. Детальное описание нод

### 4.1 Start (старт)
**Описание:** Точка входа в схему.

**Конфигурация:**
```json
{
  "type": "start",
  "id": "start_1",
  "config": {}
}
```

**Выходы:** 1 (next)

**Особенности:**
- Может быть только одна в схеме
- Всегда первая выполняемая нода

---

### 4.2 End (конец)
**Описание:** Завершает выполнение схемы.

**Конфигурация:**
```json
{
  "type": "end",
  "id": "end_1",
  "config": {
    "success": true,  // успешное или неуспешное завершение
    "message": "Execution completed"
  }
}
```

**Выходы:** 0

**Особенности:**
- Может быть несколько в схеме
- Сохраняет финальный статус

---

### 4.3 Condition (условие)
**Описание:** Проверяет условие и направляет выполнение по одному из двух путей.

**Конфигурация:**
```json
{
  "type": "condition",
  "id": "condition_1",
  "config": {
    "expression": "{{webhook.payload.age}} > 18"
  }
}
```

**Выходы:** 2 (true, false)

**Особенности:**
- Поддержка простых выражений (>, <, ==, !=, &&, ||)
- Доступ к переменным через {{}}

---

### 4.4 HTTP Request (HTTP запрос)
**Описание:** Выполняет HTTP запрос к внешнему API.

**Конфигурация:**
```json
{
  "type": "http_request",
  "id": "http_1",
  "config": {
    "method": "POST",
    "url": "https://api.example.com/users",
    "headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer {{env.API_TOKEN}}"
    },
    "body": {
      "email": "{{user.email}}",
      "name": "John Doe"
    },
    "timeout": 30,  // секунды
    "retry": {
      "enabled": true,
      "max_attempts": 3,
      "delay": 5  // секунды между попытками
    }
  }
}
```

**Выходы:** 2 (success, error)

**Результат сохраняется в:**
```json
{
  "steps.http_1.output": {
    "status_code": 200,
    "body": {...},
    "headers": {...}
  }
}
```

---

### 4.5 Variable Set (установить переменную)
**Описание:** Создаёт или обновляет переменную в контексте выполнения.

**Конфигурация:**
```json
{
  "type": "variable_set",
  "id": "var_1",
  "config": {
    "variable": "user_age",
    "value": "{{webhook.payload.age}}"
  }
}
```

**Выходы:** 1 (next)

**Результат:**
Переменная доступна как `{{variables.user_age}}`

---

### 4.6 Sleep (задержка)
**Описание:** Приостанавливает выполнение на указанное время.

**Конфигурация:**
```json
{
  "type": "sleep",
  "id": "sleep_1",
  "config": {
    "duration": 60,  // секунды
    "unit": "seconds"  // seconds|minutes|hours|days
  }
}
```

**Выходы:** 1 (next)

**Особенности:**
- Использует библиотеку `at`
- Состояние сохраняется в БД
- После задержки выполнение продолжается автоматически

---

### 4.7 Log (логирование)
**Описание:** Выводит сообщение в лог выполнения.

**Конфигурация:**
```json
{
  "type": "log",
  "id": "log_1",
  "config": {
    "level": "info",  // debug|info|warn|error
    "message": "Processing user {{user.email}}"
  }
}
```

**Выходы:** 1 (next)

**Результат:**
Лог доступен в `execution_steps` таблице

---

### 4.8 Math Operation (математика)
**Описание:** Выполняет математические операции.

**Конфигурация:**
```json
{
  "type": "math",
  "id": "math_1",
  "config": {
    "operation": "add",  // add|subtract|multiply|divide|modulo
    "operand1": "{{variables.price}}",
    "operand2": 10,
    "result_variable": "final_price"
  }
}
```

**Выходы:** 1 (next)

---

### 4.9 RabbitMQ Publish (публикация в очередь)
**Описание:** Отправляет сообщение в RabbitMQ.

**Конфигурация:**
```json
{
  "type": "rabbitmq_publish",
  "id": "rmq_1",
  "config": {
    "queue": "notifications",
    "exchange": "",  // optional
    "message": {
      "user_id": "{{variables.user_id}}",
      "text": "Welcome!"
    }
  }
}
```

**Выходы:** 2 (success, error)

---

### 4.10 SubSchema (вызов другой схемы) - ЗАГЛУШКА
**Описание:** Вызывает другую схему как подпроцесс.

**Конфигурация:**
```json
{
  "type": "sub_schema",
  "id": "sub_1",
  "config": {
    "schema_id": "uuid-другой-схемы",
    "input_mapping": {
      "param1": "{{variables.value1}}"
    }
  }
}
```

**Выходы:** 2 (success, error)

**Статус:** НЕ РЕАЛИЗОВАНО в MVP. При выполнении возвращает ошибку.

---

## 5. Интерполяция переменных

### 5.1 Синтаксис
```
{{path.to.variable}}
```

### 5.2 Доступные пути
- `webhook.payload.*` - данные от webhook
- `user.email` - email пользователя
- `execution.id` - ID выполнения
- `steps.<node_id>.output.*` - результаты предыдущих шагов
- `variables.*` - переменные, созданные через Variable Set
- `env.*` - переменные окружения (опционально)

### 5.3 Примеры
```json
"url": "https://api.example.com/users/{{webhook.payload.user_id}}"
"message": "Hello, {{user.email}}! Order #{{steps.create_order.output.order_id}}"
```

## 6. Обработка ошибок в нодах

### 6.1 Типы ошибок
- **Temporary** - можно retry (сеть, таймаут)
- **Permanent** - нельзя retry (валидация, 404)

### 6.2 Retry стратегия
```json
"retry": {
  "enabled": true,
  "max_attempts": 3,
  "delay": 5,  // секунды
  "backoff": "linear"  // linear|exponential
}
```

### 6.3 Error output
Если нода имеет выход "error", то при ошибке переходим по нему.
Если выхода нет - выполнение останавливается со статусом "failed".

## 7. Валидация нод

### 7.1 На уровне схемы (Frontend)
- Обязательные поля заполнены
- Корректный формат (URL, email, etc)
- Есть хотя бы один путь от Start до End

### 7.2 На уровне выполнения (Backend)
- Переменные существуют
- Типы данных корректны
- Циклические зависимости отсутствуют

## 8. Приоритет реализации (MVP)

### Фаза 1 (критичные для MVP):
1. Start
2. End
3. Log
4. Variable Set
5. HTTP Request
6. Condition

### Фаза 2 (расширение):
7. Sleep
8. Math Operation
9. RabbitMQ Publish

### Фаза 3 (будущее):
10. SubSchema
11. Loop
12. Switch
13. Database Query
14. JSON Operations

## 9. Открытые вопросы
- Нужна ли нода для отправки email?
- Нода для работы с файлами?
- Кастомные ноды от пользователей (плагины)?
- Как визуально различать типы нод в UI?
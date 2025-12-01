# ТЗ: Infrastructure

## 1. Общее описание
Инфраструктура для разработки и деплоя проекта "Схема".

## 2. Окружения

### 2.1 Development (локальная разработка)
- Docker Compose
- Hot reload для frontend/backend
- Моковые данные

### 2.2 Production
- Kubernetes (в перспективе)
- High availability
- Автоскейлинг

## 3. Docker Compose (для разработки)

### 3.1 Сервисы
```yaml
services:
  postgres:
    image: postgres:16
    # схема БД, миграции
  
  rabbitmq:
    image: rabbitmq:3-management
    # UI на :15672
  
  backend-api:
    build: ./backend
    # REST API
  
  backend-worker:
    build: ./backend
    # Execution Engine
  
  frontend:
    build: ./frontend
    # React dev server
```

### 3.2 Volumes
- PostgreSQL data
- RabbitMQ data
- SQLite для `at` библиотеки

### 3.3 Networks
- Внутренняя сеть для сервисов
- Проброс портов для локального доступа

## 4. База данных

### 4.1 Структура таблиц
```sql
-- Схемы
CREATE TABLE schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,  -- JSON с нодами и рёбрами
    status VARCHAR(50) DEFAULT 'draft',  -- draft|active|archived
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Выполнения
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID REFERENCES schemas(id),
    status VARCHAR(50) DEFAULT 'pending',  -- pending|running|paused|completed|failed
    trigger_type VARCHAR(50),  -- manual|webhook|scheduler
    trigger_payload JSONB,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_by VARCHAR(255),
    error TEXT
);

-- Состояние выполнения
CREATE TABLE execution_state (
    execution_id UUID PRIMARY KEY REFERENCES executions(id),
    current_node_id VARCHAR(255),
    context JSONB,  -- переменные схемы
    updated_at TIMESTAMP DEFAULT NOW()
);

-- История шагов
CREATE TABLE execution_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES executions(id),
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    input JSONB,
    output JSONB,
    status VARCHAR(50),  -- success|failed|skipped
    error TEXT,
    started_at TIMESTAMP DEFAULT NOW(),
    finished_at TIMESTAMP,
    INDEX idx_execution_steps_execution_id (execution_id)
);

-- Пользователи (заглушка на будущее)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4.2 Индексы
```sql
CREATE INDEX idx_schemas_status ON schemas(status);
CREATE INDEX idx_executions_schema_id ON executions(schema_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_execution_steps_execution_id ON execution_steps(execution_id);
```

### 4.3 Миграции
- Использовать golang-migrate или similar
- Версионирование схемы БД
- Rollback возможность

## 5. RabbitMQ

### 5.1 Очереди
```
schema_execution_queue     - основная очередь задач
schema_execution_dlq       - Dead Letter Queue
```

### 5.2 Настройки
- Durable queues
- Message persistence
- Prefetch count для воркеров
- TTL для сообщений (опционально)

### 5.3 Exchanges
- Direct exchange для основной очереди
- Retry механизм через delayed exchange (или через `at`)

## 6. Библиотека `at` для отложенных задач

### 6.1 Конфигурация
- SQLite файл для персистентности
- Callback через RabbitMQ

### 6.2 Использование
```go
at.Schedule(duration, func() {
    // Отправить задачу обратно в RabbitMQ
})
```

## 7. Логирование

### 7.1 Формат
- Структурированные JSON логи
- Уровни: DEBUG, INFO, WARN, ERROR

### 7.2 Что логировать
- Старт/конец выполнения схемы
- Каждый шаг выполнения
- Ошибки с stack trace
- API запросы (request/response)

### 7.3 Хранение
- Stdout/Stderr (для Docker)
- Rotation в продакшене

## 8. Мониторинг (для будущего)
- Prometheus метрики
- Grafana дашборды
- Healthcheck endpoints

## 9. CI/CD (для будущего)
- GitHub Actions / GitLab CI
- Автоматические тесты
- Docker image build
- Deploy в Kubernetes

## 10. Переменные окружения

### 10.1 Backend
```env
DATABASE_URL=postgresql://user:pass@localhost:5432/schema
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
LOG_LEVEL=debug
PORT=8080
```

### 10.2 Frontend
```env
REACT_APP_API_URL=http://localhost:8080/api
```

## 11. Безопасность
- Не хардкодить пароли
- Использовать secrets в production
- HTTPS в продакшене
- Rate limiting (опционально)

## 12. Открытые вопросы
- Backup стратегия для БД?
- CDN для frontend статики?
- Какой Kubernetes дистрибутив использовать?
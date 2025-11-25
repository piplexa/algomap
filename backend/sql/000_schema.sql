-- =====================================================
-- AlgoMap Database Schema 
-- Initial migration: Core tables for schema execution
-- =====================================================

-- Включаем расширение для UUID (используется только в executions)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создаём схему main (если не существует)
CREATE SCHEMA IF NOT EXISTS main;

-- =====================================================
-- СПРАВОЧНИКИ (DICTIONARIES)
-- =====================================================

-- Справочник статусов схем
CREATE TABLE main.dict_schema_status (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

COMMENT ON TABLE main.dict_schema_status IS 'Справочник статусов схем';

INSERT INTO main.dict_schema_status (id, name, description) VALUES
    (1, 'draft', 'Черновик - схема в разработке'),
    (2, 'active', 'Активна - схема работает'),
    (3, 'archived', 'Архив - схема устарела');

-- Справочник статусов выполнения
CREATE TABLE main.dict_execution_status (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

COMMENT ON TABLE main.dict_execution_status IS 'Справочник статусов выполнения';

INSERT INTO main.dict_execution_status (id, name, description) VALUES
    (1, 'pending', 'В очереди - ожидает выполнения'),
    (2, 'running', 'Выполняется'),
    (3, 'paused', 'На паузе'),
    (4, 'completed', 'Завершено успешно'),
    (5, 'failed', 'Завершено с ошибкой'),
    (6, 'stopped', 'Остановлено пользователем');

-- Справочник типов триггеров
CREATE TABLE main.dict_trigger_type (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

COMMENT ON TABLE main.dict_trigger_type IS 'Справочник типов триггеров запуска';

INSERT INTO main.dict_trigger_type (id, name, description) VALUES
    (1, 'manual', 'Ручной запуск'),
    (2, 'webhook', 'Запуск через webhook'),
    (3, 'scheduler', 'Запуск по расписанию'),
    (4, 'api', 'Запуск через API');

-- Справочник статусов шагов
CREATE TABLE main.dict_step_status (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

COMMENT ON TABLE main.dict_step_status IS 'Справочник статусов выполнения шага';

INSERT INTO main.dict_step_status (id, name, description) VALUES
    (1, 'success', 'Выполнено успешно'),
    (2, 'failed', 'Выполнено с ошибкой'),
    (3, 'skipped', 'Пропущено');

-- =====================================================
-- ТАБЛИЦА: users
-- Пользователи системы
-- =====================================================
CREATE TABLE main.users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE main.users IS 'Пользователи системы';
COMMENT ON COLUMN main.users.email IS 'Email пользователя (уникальный)';

-- =====================================================
-- ТАБЛИЦА: schemas
-- Определения схем (блок-схемы)
-- =====================================================
CREATE TABLE main.schemas (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- JSON определение схемы (ноды + рёбра)
    definition JSONB NOT NULL,
    
    -- Статус схемы (foreign key на справочник)
    id_status SMALLINT NOT NULL DEFAULT 1 REFERENCES main.dict_schema_status(id),
    
    -- Владелец
    created_by BIGINT NOT NULL REFERENCES main.users(id),
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE main.schemas IS 'Схемы автоматизации (визуальные блок-схемы)';
COMMENT ON COLUMN main.schemas.definition IS 'JSONB с нодами и рёбрами';
COMMENT ON COLUMN main.schemas.id_status IS '1=draft, 2=active, 3=archived';

-- Индексы для schemas
CREATE INDEX idx_schemas_id_status ON main.schemas(id_status);
CREATE INDEX idx_schemas_created_by ON main.schemas(created_by);
CREATE INDEX idx_schemas_created_at ON main.schemas(created_at DESC);

-- GIN индекс для поиска по JSONB
CREATE INDEX idx_schemas_definition_gin ON main.schemas USING GIN(definition);

-- =====================================================
-- ТАБЛИЦА: executions
-- История выполнений схем
-- =====================================================
CREATE TABLE main.executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Ссылка на схему
    schema_id BIGINT NOT NULL REFERENCES main.schemas(id),
    
    -- Статус выполнения (foreign key на справочник)
    id_status SMALLINT NOT NULL DEFAULT 1 REFERENCES main.dict_execution_status(id),
    
    -- Как была запущена схема (foreign key на справочник)
    id_trigger_type SMALLINT NOT NULL REFERENCES main.dict_trigger_type(id),
    
    -- Payload от триггера (webhook data, manual params, etc)
    trigger_payload JSONB,
    
    -- Текущий шаг выполнения (node_id из схемы)
    current_step_id VARCHAR(255),
    
    -- Временные метки
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Кто запустил
    created_by BIGINT NOT NULL REFERENCES main.users(id),
    
    -- Ошибка (если есть)
    error TEXT
);

COMMENT ON TABLE main.executions IS 'История выполнений схем';
COMMENT ON COLUMN main.executions.id IS 'UUID для webhook - непредсказуемость';
COMMENT ON COLUMN main.executions.id_status IS '1=pending, 2=running, 3=paused, 4=completed, 5=failed, 6=stopped';
COMMENT ON COLUMN main.executions.id_trigger_type IS '1=manual, 2=webhook, 3=scheduler, 4=api';
COMMENT ON COLUMN main.executions.trigger_payload IS 'Входные данные при запуске (например, webhook payload)';
COMMENT ON COLUMN main.executions.current_step_id IS 'ID ноды, которая сейчас выполняется (для отображения прогресса)';

-- Индексы для executions
CREATE INDEX idx_executions_schema_id ON main.executions(schema_id);
CREATE INDEX idx_executions_id_status ON main.executions(id_status);
CREATE INDEX idx_executions_created_at ON main.executions(created_at DESC);
CREATE INDEX idx_executions_started_at ON main.executions(started_at DESC) WHERE started_at IS NOT NULL;
CREATE INDEX idx_executions_created_by ON main.executions(created_by);

-- Составной индекс для поиска активных выполнений конкретной схемы
CREATE INDEX idx_executions_schema_status ON main.executions(schema_id, id_status);

-- =====================================================
-- ТАБЛИЦА: execution_state
-- Текущее состояние выполнения (для resume)
-- =====================================================
CREATE TABLE main.execution_state (
    execution_id UUID PRIMARY KEY REFERENCES main.executions(id),
    
    -- ID текущей ноды, на которой остановились
    current_node_id VARCHAR(255) NOT NULL,
    
    -- Контекст выполнения (переменные, результаты предыдущих шагов)
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Временная метка последнего обновления
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE main.execution_state IS 'Текущее состояние выполнения схемы (для pause/resume)';
COMMENT ON COLUMN main.execution_state.current_node_id IS 'ID ноды, которая сейчас выполняется или следующая для выполнения';
COMMENT ON COLUMN main.execution_state.context IS 'Переменные и данные, доступные в схеме: webhook.payload, steps.*.output, variables.*';

-- GIN индекс для поиска по context
CREATE INDEX idx_execution_state_context_gin ON main.execution_state USING GIN(context);

-- =====================================================
-- ТАБЛИЦА: execution_steps
-- Детальная история выполнения каждой ноды
-- =====================================================
CREATE TABLE main.execution_steps (
    id BIGSERIAL PRIMARY KEY,
    
    -- Ссылка на выполнение
    execution_id UUID NOT NULL REFERENCES main.executions(id),
    
    -- Информация о ноде
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    
    -- Предыдущая и следующая ноды (для понимания потока выполнения)
    prev_node_id VARCHAR(255),
    next_node_id VARCHAR(255),
    
    -- Входные данные для ноды
    input JSONB,
    
    -- Результат выполнения ноды
    output JSONB,
    
    -- Статус выполнения этого шага (foreign key на справочник)
    id_status SMALLINT NOT NULL REFERENCES main.dict_step_status(id),
    
    -- Ошибка (если есть)
    error TEXT,
    
    -- Временные метки
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP
);

COMMENT ON TABLE main.execution_steps IS 'Детальная история выполнения каждого шага (ноды)';
COMMENT ON COLUMN main.execution_steps.node_id IS 'ID ноды в схеме (например, "http_1", "condition_2")';
COMMENT ON COLUMN main.execution_steps.node_type IS 'Тип ноды (http_request, condition, sleep, etc)';
COMMENT ON COLUMN main.execution_steps.prev_node_id IS 'ID предыдущей ноды (откуда пришли)';
COMMENT ON COLUMN main.execution_steps.next_node_id IS 'ID следующей ноды (куда пошли)';
COMMENT ON COLUMN main.execution_steps.input IS 'Входные параметры для ноды';
COMMENT ON COLUMN main.execution_steps.output IS 'Результат выполнения ноды';
COMMENT ON COLUMN main.execution_steps.id_status IS '1=success, 2=failed, 3=skipped';

-- Индексы для execution_steps
CREATE INDEX idx_execution_steps_execution_id ON main.execution_steps(execution_id);
CREATE INDEX idx_execution_steps_started_at ON main.execution_steps(started_at DESC);
CREATE INDEX idx_execution_steps_id_status ON main.execution_steps(id_status);

-- Составной индекс для поиска шагов конкретного выполнения
CREATE INDEX idx_execution_steps_execution_started ON main.execution_steps(execution_id, started_at DESC);

-- =====================================================
-- ТАБЛИЦА: webhook_configs (опционально для MVP)
-- Конфигурация webhook endpoint'ов для схем
-- =====================================================
CREATE TABLE main.webhook_configs (
    id BIGSERIAL PRIMARY KEY,
    
    -- Ссылка на схему
    schema_id BIGINT NOT NULL REFERENCES main.schemas(id),
    
    -- Уникальный токен для webhook URL
    webhook_token VARCHAR(255) NOT NULL UNIQUE,
    
    -- Активен ли webhook
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE main.webhook_configs IS 'Конфигурация webhook для запуска схем';
COMMENT ON COLUMN main.webhook_configs.webhook_token IS 'Уникальный токен для URL: /webhook/{webhook_token}';

-- Индексы для webhook_configs
CREATE INDEX idx_webhook_configs_schema_id ON main.webhook_configs(schema_id);
CREATE UNIQUE INDEX idx_webhook_configs_token ON main.webhook_configs(webhook_token);

-- =====================================================
-- ПРИМЕЧАНИЯ:
-- 
-- 1. BIGINT для ID - компактность и читаемость (кроме executions)
-- 2. UUID только в executions - для webhook непредсказуемость
-- 3. JSONB вместо JSON - быстрее и можно индексировать
-- 4. GIN индексы для быстрого поиска в JSONB
-- 5. Все временные метки в UTC (NOW() в PostgreSQL = UTC при правильной настройке)
-- 6. Справочники вместо CHECK constraints - экономия памяти и гибкость
-- 7. НЕТ CASCADE - безопасность превыше удобства
-- 8. Схема main - чистое пространство для бизнес-таблиц
-- 9. prev_node_id / next_node_id - для визуализации потока
-- 10. current_step_id в executions - для отображения прогресса
-- =====================================================
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Execution представляет запуск схемы
type Execution struct {
	ID              uuid.UUID       `json:"id"`
	SchemaID        int64           `json:"schema_id"`
	Status          int16           `json:"status"`           // 1=pending, 2=running, 3=paused, 4=completed, 5=failed, 6=stopped
	TriggerType     int16           `json:"trigger_type"`     // 1=manual, 2=webhook, 3=scheduler, 4=api
	TriggerPayload  json.RawMessage `json:"trigger_payload"`  // Данные от триггера
	CurrentStepID   *string         `json:"current_step_id"`  // ID текущей ноды
	StartedAt       *time.Time      `json:"started_at"`
	FinishedAt      *time.Time      `json:"finished_at"`
	CreatedAt       time.Time       `json:"created_at"`
	CreatedBy       int64           `json:"created_by"`
	Error           *string         `json:"error"`
}

// ExecutionState представляет текущее состояние выполнения
type ExecutionState struct {
	ExecutionID   uuid.UUID       `json:"execution_id"`
	CurrentNodeID string          `json:"current_node_id"`
	Context       json.RawMessage `json:"context"` // Переменные схемы
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ExecutionStep представляет один шаг выполнения (нода)
type ExecutionStep struct {
	ID           int64           `json:"id"`
	ExecutionID  uuid.UUID       `json:"execution_id"`
	NodeID       string          `json:"node_id"`
	NodeType     string          `json:"node_type"`
	PrevNodeID   *string         `json:"prev_node_id"`
	NextNodeID   *string         `json:"next_node_id"`
	Input        json.RawMessage `json:"input"`
	Output       json.RawMessage `json:"output"`
	Status       int16           `json:"status"` // 1=success, 2=failed, 3=skipped
	Error        *string         `json:"error"`
	StartedAt    time.Time       `json:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at"`
}

// TODO: Добавить методы и бизнес-логику для работы с Execution
// TODO: Реализовать в worker'е
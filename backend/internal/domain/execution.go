package domain

import (
	"time"
)

// Execution_view представляет представление выполнения схемы с агрегированными данными
// для просмотра списка выполнений
type Execution_view struct {
	ID			   	string  	`json:"id" db:"id"`
	SchemaID       	int64   	`json:"schema_id" db:"schema_id"`
	CreatedAt      	*time.Time	`json:"created_at" db:"created_at"`
	FinishedAt     	*time.Time 	`json:"finished_at" db:"finished_at"`
	CntExecutedSteps int64    	`json:"cnt_executed_steps" db:"cnt_executed_steps"`
	Duration       	*float64   	`json:"duration" db:"duration"`
	Status			string		`json:"status" db:"status"`
}

// Execution представляет запуск схемы
type Execution struct {
	ID              string                 `json:"id" db:"id"`
	SchemaID        int64                  `json:"schema_id" db:"schema_id"`
	StatusID        int16                  `json:"status_id" db:"id_status"`
	StatusName      string                 `json:"status_name,omitempty"` // для JOIN с dict_execution_status
	TriggerTypeID   int16                  `json:"trigger_type_id" db:"id_trigger_type"`
	TriggerTypeName string                 `json:"trigger_type_name,omitempty"` // для JOIN
	TriggerPayload  map[string]interface{} `json:"trigger_payload,omitempty" db:"trigger_payload"`
	CurrentStepID   *string                `json:"current_step_id,omitempty" db:"current_step_id"`
	StartedAt       *time.Time             `json:"started_at,omitempty" db:"started_at"`
	FinishedAt      *time.Time             `json:"finished_at,omitempty" db:"finished_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	CreatedBy       int64                  `json:"created_by" db:"created_by"`
	Error           *string                `json:"error,omitempty" db:"error"`
}

// ExecutionState представляет текущее состояние выполнения
type ExecutionState struct {
	ExecutionID   string                 `json:"execution_id" db:"execution_id"`
	CurrentNodeID string                 `json:"current_node_id" db:"current_node_id"`
	Context       map[string]interface{} `json:"context" db:"context"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// ExecutionStep представляет один шаг выполнения
type ExecutionStep struct {
	ID           int64                  `json:"id" db:"id"`
	ExecutionID  string                 `json:"execution_id" db:"execution_id"`
	NodeID       string                 `json:"node_id" db:"node_id"`
	NodeType     string                 `json:"node_type" db:"node_type"`
	PrevNodeID   *string                `json:"prev_node_id,omitempty" db:"prev_node_id"`
	NextNodeID   *string                `json:"next_node_id,omitempty" db:"next_node_id"`
	Input        map[string]interface{} `json:"input,omitempty" db:"input"`
	Output       map[string]interface{} `json:"output,omitempty" db:"output"`
	StatusID     int16                  `json:"status_id" db:"id_status"`
	StatusName   string                 `json:"status_name,omitempty"`
	Error        *string                `json:"error,omitempty" db:"error"`
	StartedAt    time.Time              `json:"started_at" db:"started_at"`
	FinishedAt   *time.Time             `json:"finished_at,omitempty" db:"finished_at"`
}

// Константы для статусов выполнения
const (
	ExecutionStatusPending   int16 = 1
	ExecutionStatusRunning   int16 = 2
	ExecutionStatusPaused    int16 = 3
	ExecutionStatusCompleted int16 = 4
	ExecutionStatusFailed    int16 = 5
	ExecutionStatusStopped   int16 = 6
)

// Константы для типов триггеров
const (
	TriggerTypeManual    int16 = 1
	TriggerTypeWebhook   int16 = 2
	TriggerTypeScheduler int16 = 3
	TriggerTypeAPI       int16 = 4
)

// Константы для статусов шагов
const (
	StepStatusSuccess int16 = 1
	StepStatusFailed  int16 = 2
	StepStatusSkipped int16 = 3
)

// CreateExecutionResponse - ответ при создании execution
type CreateExecutionResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}
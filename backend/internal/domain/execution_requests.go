package domain

import "encoding/json"

// CreateExecutionRequest - запрос на создание execution
type CreateExecutionRequest struct {
	SchemaID       int64           `json:"schema_id" binding:"required"`
	TriggerPayload json.RawMessage `json:"trigger_payload,omitempty"`
	DebugMode      bool            `json:"debug_mode,omitempty"`
}

// Константы для статусов схем (для проверки в handlers)
const (
	SchemaStatusDraft    int16 = 1
	SchemaStatusActive   int16 = 2
	SchemaStatusArchived int16 = 3
)
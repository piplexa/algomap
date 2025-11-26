package domain

import (
	"encoding/json"
	"time"
)

// Schema представляет схему автоматизации
type Schema struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"` // JSONB с нодами и рёбрами
	Status      int16           `json:"status"`     // 1=draft, 2=active, 3=archived
	CreatedBy   int64           `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// SchemaDefinition - структура для валидации definition (опционально)
type SchemaDefinition struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node представляет ноду в схеме
type Node struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

// Edge представляет связь между нодами
type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"` // ID ноды-источника
	Target string `json:"target"` // ID ноды-назначения
	Label  string `json:"label"`  // Метка (например, "true", "false", "next")
}

// CreateSchemaRequest - запрос на создание схемы
type CreateSchemaRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"`
}

// UpdateSchemaRequest - запрос на обновление схемы
type UpdateSchemaRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Definition  *json.RawMessage `json:"definition,omitempty"`
	Status      *int16           `json:"status,omitempty"`
}
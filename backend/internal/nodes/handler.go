package nodes

import (
	"context"
	"encoding/json"
	"time"
)

// ExecutionStatus статусы выполнения ноды
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSleep   = "sleep"
)

// Node представляет ноду в схеме
type Node struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Config json.RawMessage `json:"config"`
}

// NodeResult результат выполнения ноды
type NodeResult struct {
	Output     map[string]interface{} `json:"output"`
	NextNodeID *string                `json:"next_node_id,omitempty"`
	Status     string                 `json:"status"`
	Error      *string                `json:"error,omitempty"`
	SleepUntil *time.Time             `json:"sleep_until,omitempty"`
}

// ExecutionContext контекст выполнения схемы
type ExecutionContext struct {
	Webhook   map[string]interface{} `json:"webhook,omitempty"`
	User      map[string]interface{} `json:"user"`
	Execution map[string]interface{} `json:"execution"`
	Steps     map[string]StepOutput  `json:"steps"`
	Variables map[string]interface{} `json:"variables"`
}

// StepOutput результат выполнения шага
type StepOutput struct {
	Output map[string]interface{} `json:"output"`
}

// NodeHandler интерфейс обработчика ноды
type NodeHandler interface {
	// Execute выполняет ноду и возвращает результат
	Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error)
}

// HandlerRegistry реестр обработчиков нод
type HandlerRegistry struct {
	handlers map[string]NodeHandler
}

// NewHandlerRegistry создаёт новый реестр
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]NodeHandler),
	}
}

// Register регистрирует обработчик для типа ноды
func (r *HandlerRegistry) Register(nodeType string, handler NodeHandler) {
	r.handlers[nodeType] = handler
}

// Get возвращает обработчик для типа ноды
func (r *HandlerRegistry) Get(nodeType string) (NodeHandler, bool) {
	handler, ok := r.handlers[nodeType]
	return handler, ok
}

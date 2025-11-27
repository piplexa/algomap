package nodes

import (
	"context"
)

// StartHandler обработчик стартовой ноды
type StartHandler struct{}

// NewStartHandler создаёт новый StartHandler
func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

// Execute выполняет start ноду
// Start нода просто возвращает success и позволяет двигаться дальше
func (h *StartHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	return &NodeResult{
		Output: map[string]interface{}{
			"started": true,
		},
		Status: StatusSuccess,
	}, nil
}

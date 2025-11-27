package nodes

import (
	"context"
)

// EndHandler обработчик завершающей ноды
type EndHandler struct{}

// NewEndHandler создаёт новый EndHandler
func NewEndHandler() *EndHandler {
	return &EndHandler{}
}

// Execute выполняет end ноду
// End нода просто фиксирует завершение выполнения
func (h *EndHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	return &NodeResult{
		Output: map[string]interface{}{
			"completed": true,
		},
		Status: StatusSuccess,
		// NextNodeID nil означает, что это последняя нода
	}, nil
}

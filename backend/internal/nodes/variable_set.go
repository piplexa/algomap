package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// VariableSetConfig конфигурация variable_set ноды
type VariableSetConfig struct {
	Variable string      `json:"variable"` // ← Исправлено с variable_name
	Value    interface{} `json:"value"`    // ← Исправлено с variable_value
}

// VariableSetHandler обработчик ноды установки переменной
type VariableSetHandler struct{}

// NewVariableSetHandler создаёт новый VariableSetHandler
func NewVariableSetHandler() *VariableSetHandler {
	return &VariableSetHandler{}
}

// Execute выполняет variable_set ноду
func (h *VariableSetHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext, preNextIdNode *string) (*NodeResult, error) {
	var config VariableSetConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse variable_set config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	if config.Variable == "" { // ← Исправлено с VariableName
		errMsg := "variable is required"
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// TODO: Интерполировать значение если это строка с {{...}}
	value := config.Value // ← Исправлено с VariableValue

	// Устанавливаем переменную в контекст
	if execCtx.Variables == nil {
		execCtx.Variables = make(map[string]interface{})
	}
	execCtx.Variables[config.Variable] = value // ← Исправлено с VariableName

	return &NodeResult{
		Output: map[string]interface{}{
			"variable": config.Variable, // ← Исправлено
			"value":    value,           // ← Исправлено
		},
		Status: StatusSuccess,
	}, nil
}
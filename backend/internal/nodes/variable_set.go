package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// VariableSetConfig конфигурация variable_set ноды
type VariableSetConfig struct {
	VariableName  string      `json:"variable_name"`
	VariableValue interface{} `json:"variable_value"`
}

// VariableSetHandler обработчик ноды установки переменной
type VariableSetHandler struct{}

// NewVariableSetHandler создаёт новый VariableSetHandler
func NewVariableSetHandler() *VariableSetHandler {
	return &VariableSetHandler{}
}

// Execute выполняет variable_set ноду
func (h *VariableSetHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config VariableSetConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse variable_set config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	if config.VariableName == "" {
		errMsg := "variable_name is required"
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// TODO: Интерполировать значение если это строка с {{...}}
	value := config.VariableValue

	// Устанавливаем переменную в контекст
	if execCtx.Variables == nil {
		execCtx.Variables = make(map[string]interface{})
	}
	execCtx.Variables[config.VariableName] = value

	return &NodeResult{
		Output: map[string]interface{}{
			"variable_name":  config.VariableName,
			"variable_value": value,
		},
		Status: StatusSuccess,
	}, nil
}

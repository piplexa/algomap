package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// MathConfig конфигурация math ноды
type MathConfig struct {
	Operation string      `json:"operation"` // add, subtract, multiply, divide
	Left      interface{} `json:"left"`
	Right     interface{} `json:"right"`
}

// MathHandler обработчик математических операций
type MathHandler struct{}

// NewMathHandler создаёт новый MathHandler
func NewMathHandler() *MathHandler {
	return &MathHandler{}
}

// Execute выполняет math ноду
func (h *MathHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config MathConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse math config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// TODO: Интерполировать left и right если это {{...}}

	// Конвертируем в float64
	left, err := toFloat64(config.Left)
	if err != nil {
		errMsg := fmt.Sprintf("invalid left operand: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	right, err := toFloat64(config.Right)
	if err != nil {
		errMsg := fmt.Sprintf("invalid right operand: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// Выполняем операцию
	var result float64
	switch config.Operation {
	case "add":
		result = left + right
	case "subtract":
		result = left - right
	case "multiply":
		result = left * right
	case "divide":
		if right == 0 {
			errMsg := "division by zero"
			return &NodeResult{
				Status: StatusFailed,
				Error:  &errMsg,
			}, nil
		}
		result = left / right
	default:
		errMsg := fmt.Sprintf("unknown operation: %s", config.Operation)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	return &NodeResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": config.Operation,
			"left":      left,
			"right":     right,
		},
		Status: StatusSuccess,
	}, nil
}

// toFloat64 конвертирует interface{} в float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

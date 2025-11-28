package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// MathConfig конфигурация math ноды
type MathConfig struct {
	Operation      string      `json:"operation"`       // add, subtract, multiply, divide
	Operand1       interface{} `json:"operand1"`        // ← Исправлено с left
	Operand2       interface{} `json:"operand2"`        // ← Исправлено с right
	ResultVariable string      `json:"result_variable"` // ← Добавлено
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

	// TODO: Интерполировать operand1 и operand2 если это строки с {{...}}
	// Пока просто пытаемся получить значения из переменных если это строки
	operand1 := resolveValue(config.Operand1, execCtx)
	operand2 := resolveValue(config.Operand2, execCtx)

	// Конвертируем в float64
	left, err := toFloat64(operand1)
	if err != nil {
		errMsg := fmt.Sprintf("invalid operand1: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	right, err := toFloat64(operand2)
	if err != nil {
		errMsg := fmt.Sprintf("invalid operand2: %v", err)
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

	// Сохраняем результат в переменную если указано
	if config.ResultVariable != "" {
		if execCtx.Variables == nil {
			execCtx.Variables = make(map[string]interface{})
		}
		execCtx.Variables[config.ResultVariable] = result
	}

	return &NodeResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": config.Operation,
			"operand1":  left,
			"operand2":  right,
		},
		Status: StatusSuccess,
	}, nil
}

// resolveValue пытается получить значение из переменных если это строка
func resolveValue(v interface{}, ctx *ExecutionContext) interface{} {
	// Если это строка - проверяем, не имя ли это переменной
	if str, ok := v.(string); ok {
		// Проверяем есть ли такая переменная
		if val, exists := ctx.Variables[str]; exists {
			return val
		}
	}
	return v
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
	case string:
		// Пытаемся распарсить строку как число
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string '%s' to float64: %w", val, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
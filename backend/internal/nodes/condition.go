package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ConditionConfig конфигурация condition ноды
type ConditionConfig struct {
	Expression string `json:"expression"` // Например: "{{x}} > 10"
}

// ConditionHandler обработчик условий
type ConditionHandler struct{}

// NewConditionHandler создаёт новый ConditionHandler
func NewConditionHandler() *ConditionHandler {
	return &ConditionHandler{}
}

// Execute выполняет condition ноду
func (h *ConditionHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext, preNextIdNode *string) (*NodeResult, error) {
	var config ConditionConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse condition config: %v", err)
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	if config.Expression == "" {
		errMsg := "expression is required"
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	// 1. Интерполируем переменные в выражении
	interpolated := InterpolateString(config.Expression, execCtx)

	// 2. Вычисляем выражение
	result, err := evaluateExpression(interpolated)
	if err != nil {
		errMsg := fmt.Sprintf("failed to evaluate expression '%s': %v", interpolated, err)
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	// 3. Определяем exitHandle на основе результата
	exitHandle := "false"
	if result {
		exitHandle = "true"
	}

	return &NodeResult{
		Output: map[string]interface{}{
			"expression":   config.Expression,
			"interpolated": interpolated,
			"result":       result,
		},
		Status:     StatusSuccess,
		ExitHandle: exitHandle,
	}, nil
}

// evaluateExpression вычисляет простое логическое выражение
// Поддерживает: >, <, >=, <=, ==, !=, &&, ||
func evaluateExpression(expr string) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Обрабатываем логические операторы (И, ИЛИ)
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			result, err := evaluateExpression(strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if result {
				return true, nil // Хотя бы одно true
			}
		}
		return false, nil
	}

	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			result, err := evaluateExpression(strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil // Хотя бы одно false
			}
		}
		return true, nil
	}

	// Обрабатываем операторы сравнения
	operators := []string{">=", "<=", "==", "!=", ">", "<"}
	
	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			if len(parts) != 2 {
				return false, fmt.Errorf("invalid expression: %s", expr)
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			return compareValues(left, right, op)
		}
	}

	// Если нет операторов - пытаемся интерпретировать как boolean
	return parseBoolean(expr)
}

// compareValues сравнивает два значения
func compareValues(left, right, operator string) (bool, error) {
	// Пытаемся преобразовать в числа
	leftNum, leftIsNum := parseNumber(left)
	rightNum, rightIsNum := parseNumber(right)

	if leftIsNum && rightIsNum {
		return compareNumbers(leftNum, rightNum, operator), nil
	}

	// Если не числа - сравниваем как строки
	return compareStrings(left, right, operator), nil
}

// parseNumber пытается распарсить строку в число
func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	
	// Убираем кавычки если есть
	s = strings.Trim(s, `"'`)
	
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return num, true
}

// compareNumbers сравнивает числа
func compareNumbers(left, right float64, operator string) bool {
	switch operator {
	case ">":
		return left > right
	case "<":
		return left < right
	case ">=":
		return left >= right
	case "<=":
		return left <= right
	case "==":
		return left == right
	case "!=":
		return left != right
	default:
		return false
	}
}

// compareStrings сравнивает строки
func compareStrings(left, right, operator string) bool {
	// Убираем кавычки
	left = strings.Trim(strings.TrimSpace(left), `"'`)
	right = strings.Trim(strings.TrimSpace(right), `"'`)

	switch operator {
	case "==":
		return left == right
	case "!=":
		return left != right
	case ">":
		return left > right
	case "<":
		return left < right
	case ">=":
		return left >= right
	case "<=":
		return left <= right
	default:
		return false
	}
}

// parseBoolean пытается распарсить строку как boolean
func parseBoolean(s string) (bool, error) {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	s = strings.ToLower(s)

	switch s {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no", "":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse '%s' as boolean", s)
	}
}
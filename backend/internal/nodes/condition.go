package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// ConditionConfig конфигурация condition ноды
type ConditionConfig struct {
	Expression string `json:"expression"` // Например: "{{variables.x}} > 10"
	TrueBranch string `json:"true_branch"`  // ID ноды для true
	FalseBranch string `json:"false_branch"` // ID ноды для false
}

// ConditionHandler обработчик условий
type ConditionHandler struct{}

// NewConditionHandler создаёт новый ConditionHandler
func NewConditionHandler() *ConditionHandler {
	return &ConditionHandler{}
}

// Execute выполняет condition ноду
func (h *ConditionHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config ConditionConfig
	if err := json.Unmarshal(node.Data, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse condition config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// TODO: Реализовать:
	// 1. Интерполировать переменные в expression
	//    expr := interpolateVariables(config.Expression, execCtx)
	//
	// 2. Распарсить и вычислить выражение
	//    Варианты:
	//    a) Простой парсер для базовых операций (>, <, ==, !=, >=, <=, &&, ||)
	//    b) Использовать библиотеку типа govaluate
	//    c) Написать свой DSL
	//
	//    result, err := evaluateExpression(expr)
	//
	// 3. Определить следующую ноду на основе result
	//    var nextNode string
	//    if result {
	//        nextNode = config.TrueBranch
	//    } else {
	//        nextNode = config.FalseBranch
	//    }
	//
	// 4. Вернуть результат
	//    return &NodeResult{
	//        Output: map[string]interface{}{
	//            "expression": config.Expression,
	//            "result": result,
	//            "branch": nextNode,
	//        },
	//        NextNodeID: &nextNode,
	//        Status: StatusSuccess,
	//    }, nil

	// Заглушка
	errMsg := "condition node not implemented yet"
	return &NodeResult{
		Status: StatusFailed,
		Error:  &errMsg,
	}, nil
}

// TODO: Примеры поддерживаемых выражений:
// "{{variables.x}} > 10"
// "{{variables.status}} == 'active'"
// "{{steps.http_1.output.status_code}} == 200"
// "{{variables.age}} >= 18 && {{variables.country}} == 'US'"
// "{{variables.balance}} > 100 || {{variables.is_premium}} == true"

// TODO: Реализовать evaluateExpression
// func evaluateExpression(expr string) (bool, error) {
//     // Вариант 1: Простой парсер для базовых операций
//     // Вариант 2: github.com/Knetic/govaluate
//     // Вариант 3: github.com/antonmedv/expr
//     return false, fmt.Errorf("not implemented")
// }

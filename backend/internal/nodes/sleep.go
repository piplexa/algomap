package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SleepConfig конфигурация sleep ноды
type SleepConfig struct {
	Duration int    `json:"duration"` // Длительность в секундах
	Unit     string `json:"unit"`     // seconds, minutes, hours
}

// SleepHandler обработчик ноды задержки
type SleepHandler struct{}

// NewSleepHandler создаёт новый SleepHandler
func NewSleepHandler() *SleepHandler {
	return &SleepHandler{}
}

// Execute выполняет sleep ноду
func (h *SleepHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config SleepConfig
	if err := json.Unmarshal(node.Data, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse sleep config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// Вычисляем время пробуждения
	var duration time.Duration
	switch config.Unit {
	case "seconds", "":
		duration = time.Duration(config.Duration) * time.Second
	case "minutes":
		duration = time.Duration(config.Duration) * time.Minute
	case "hours":
		duration = time.Duration(config.Duration) * time.Hour
	default:
		errMsg := fmt.Sprintf("invalid unit: %s", config.Unit)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	sleepUntil := time.Now().Add(duration)

	return &NodeResult{
		Output: map[string]interface{}{
			"sleep_until": sleepUntil,
			"duration":    duration.String(),
		},
		Status:     StatusSleep,
		SleepUntil: &sleepUntil,
	}, nil
}

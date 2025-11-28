package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// LogConfig конфигурация log ноды
type LogConfig struct {
	Message string `json:"message"`
	Level   string `json:"level"` // debug, info, warn, error
}

// LogHandler обработчик ноды логирования
type LogHandler struct {
	logger *zap.Logger
}

// NewLogHandler создаёт новый LogHandler
func NewLogHandler(logger *zap.Logger) *LogHandler {
	return &LogHandler{
		logger: logger,
	}
}

// Execute выполняет log ноду
func (h *LogHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config LogConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse log config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// Интерполируем переменные в сообщении
	message := InterpolateString(config.Message, execCtx)

	// Логируем в зависимости от уровня
	switch config.Level {
	case "debug":
		h.logger.Debug(message)
	case "warn":
		h.logger.Warn(message)
	case "error":
		h.logger.Error(message)
	default:
		h.logger.Info(message)
	}

	return &NodeResult{
		Output: map[string]interface{}{
			"message": message,
			"level":   config.Level,
		},
		Status: StatusSuccess,
	}, nil
}
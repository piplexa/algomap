package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SleepConfig конфигурация sleep ноды
type SleepConfig struct {
	Duration int    `json:"duration"` // Длительность в секундах
	Unit     string `json:"unit"`     // seconds, minutes, hours
}

// ATSchedulerTaskRequest структура запроса к AT Scheduler
type ATSchedulerTaskRequest struct {
	ExecuteAt   string                 `json:"execute_at"`
	TaskType    string                 `json:"task_type"`
	Payload     map[string]interface{} `json:"payload"`
	MaxAttempts int                    `json:"max_attempts"`
}

// SleepHandler обработчик ноды задержки
type SleepHandler struct {
	atSchedulerURL string
	urlExecution   string
	logger         *zap.Logger
}

// NewSleepHandler создаёт новый SleepHandler
func NewSleepHandler(atSchedulerURL, urlExecution string, logger *zap.Logger) *SleepHandler {
	return &SleepHandler{
		atSchedulerURL: atSchedulerURL,
		urlExecution:   urlExecution,
		logger:         logger,
	}
}

// Execute выполняет sleep ноду
func (h *SleepHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext, preNextIdNode *string) (*NodeResult, error) {
	var config SleepConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
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

	// Отправляем задачу в AT Scheduler
	if err := h.scheduleWakeUp(ctx, sleepUntil, execCtx, *preNextIdNode); err != nil {
		errMsg := fmt.Sprintf("failed to schedule wake up in AT Scheduler: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	return &NodeResult{
		Output: map[string]interface{}{
			"sleep_until": sleepUntil,
			"duration":    duration.String(),
		},
		Status:     StatusSleep,
		SleepUntil: &sleepUntil,
	}, nil
}

// scheduleWakeUp отправляет задачу в AT Scheduler для пробуждения схемы
func (h *SleepHandler) scheduleWakeUp(ctx context.Context, sleepUntil time.Time, execCtx *ExecutionContext, nodeID string) error {
	// Получаем execution ID из контекста
	executionID, ok := execCtx.Execution["id"]
	if !ok {
		return fmt.Errorf("execution id not found in context")
	}

	// Формируем URL для callback
	callbackURL := fmt.Sprintf("%s/api/executions/%v/%s/continue", h.urlExecution, executionID, nodeID)

	// Создаем запрос для AT Scheduler
	taskRequest := ATSchedulerTaskRequest{
		ExecuteAt: sleepUntil.Format(time.RFC3339),
		TaskType:  "http_callback",
		Payload: map[string]interface{}{
			"url":    callbackURL,
			"method": "POST",
			"data":   map[string]interface{}{},
		},
		MaxAttempts: 3,
	}

	// Сериализуем запрос в JSON
	requestBody, err := json.Marshal(taskRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal task request: %w", err)
	}

	// Формируем URL для AT Scheduler
	schedulerURL := fmt.Sprintf("%s/api/v1/tasks", h.atSchedulerURL)

	// Логируем параметры запроса
	h.logger.Debug("Sending wake up task to AT Scheduler",
		zap.String("scheduler_url", schedulerURL),
		zap.String("callback_url", callbackURL),
		zap.String("execute_at", sleepUntil.Format(time.RFC3339)),
		zap.Any("execution_id", executionID),
		zap.String("node_id", nodeID),
	)

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, schedulerURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to AT Scheduler: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AT Scheduler returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

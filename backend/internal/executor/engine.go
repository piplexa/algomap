package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/piplexa/algomap/internal/domain"
	"github.com/piplexa/algomap/internal/nodes"
)

// Engine движок выполнения схем
type Engine struct {
	db       *sql.DB
	logger   *zap.Logger
	registry *nodes.HandlerRegistry
	timeout  time.Duration
}

// NewEngine создаёт новый движок
func NewEngine(db *sql.DB, logger *zap.Logger, registry *nodes.HandlerRegistry) *Engine {
	return &Engine{
		db:       db,
		logger:   logger,
		registry: registry,
		timeout:  30 * time.Second, // Таймаут на выполнение одной ноды
	}
}

// ExecutionMessage сообщение из RabbitMQ
type ExecutionMessage struct {
	ExecutionID   string `json:"execution_id"`
	SchemaID      int64  `json:"schema_id"`
	CurrentNodeID string `json:"current_node_id"`
	DebugMode     bool   `json:"debug_mode"`
}

// SchemaDefinition определение схемы
type SchemaDefinition struct {
	Nodes []nodes.Node  `json:"nodes"`
	Edges []SchemaEdge  `json:"edges"`
}

// SchemaEdge связь между нодами
type SchemaEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ExecutionState состояние выполнения
type ExecutionState struct {
	ExecutionID   string                  `json:"execution_id"`
	CurrentNodeID string                  `json:"current_node_id"`
	Context       *nodes.ExecutionContext `json:"context"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

// Execute выполняет один шаг схемы и возвращает ID следующей ноды (если есть)
func (e *Engine) Execute(ctx context.Context, msg *ExecutionMessage) (*string, error) {
	e.logger.Info("executing node",
		zap.String("execution_id", msg.ExecutionID),
		zap.String("node_id", msg.CurrentNodeID),
	)

	// Создаём контекст с таймаутом
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Начинаем транзакцию
	tx, err := e.db.BeginTx(execCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Загружаем состояние выполнения
	state, err := e.loadExecutionState(execCtx, tx, msg.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load execution state: %w", err)
	}

	// Если state == nil, это первый запуск - инициализируем
	if state == nil {
		state = e.initializeState(msg)
	}

	// 2. Загружаем схему
	schema, err := e.loadSchema(execCtx, tx, msg.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	// 3. Находим ноду
	node := e.findNode(schema, msg.CurrentNodeID)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", msg.CurrentNodeID)
	}

	// 4. Получаем обработчик ноды (используем реальный тип из data.type)
	handler, ok := e.registry.Get(node.Data.Type)
	if !ok {
		return nil, fmt.Errorf("handler not found for node type: %s", node.Data.Type)
	}

	// 5. Выполняем ноду
	startedAt := time.Now()
	result, err := handler.Execute(execCtx, node, state.Context)
	finishedAt := time.Now()

	if err != nil {
		// Сохраняем ошибку
		errMsg := err.Error()
		result = &nodes.NodeResult{
			Status: nodes.StatusFailed,
			Error:  &errMsg,
		}
	}

	// 6. Сохраняем шаг в execution_steps
	if err := e.saveExecutionStep(execCtx, tx, msg.ExecutionID, node, result, startedAt, finishedAt); err != nil {
		return nil, fmt.Errorf("failed to save execution step: %w", err)
	}

	// 7. Обновляем контекст
	e.updateContext(state.Context, msg.CurrentNodeID, result)

	// 8. Определяем следующую ноду
	var nextNodeID *string
	if result.Status == nodes.StatusSuccess && node.Data.Type != domain.NodeTypeEnd {
		nextNodeID = e.findNextNode(schema, msg.CurrentNodeID)
	}

	// 9. Сохраняем обновлённое состояние
	state.CurrentNodeID = msg.CurrentNodeID
	if nextNodeID != nil {
		state.CurrentNodeID = *nextNodeID
	}
	state.UpdatedAt = time.Now()

	if err := e.saveExecutionState(execCtx, tx, state); err != nil {
		return nil, fmt.Errorf("failed to save execution state: %w", err)
	}

	// 10. Обновляем статус execution
	if err := e.updateExecutionStatus(execCtx, tx, msg, result, nextNodeID); err != nil {
		return nil, fmt.Errorf("failed to update execution status: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	e.logger.Info("node executed successfully",
		zap.String("execution_id", msg.ExecutionID),
		zap.String("node_id", msg.CurrentNodeID),
		zap.String("status", result.Status),
	)

	// Возвращаем ID следующей ноды (если есть)
	return nextNodeID, nil
}

// initializeState создаёт начальное состояние
func (e *Engine) initializeState(msg *ExecutionMessage) *ExecutionState {
	return &ExecutionState{
		ExecutionID:   msg.ExecutionID,
		CurrentNodeID: msg.CurrentNodeID,
		Context: &nodes.ExecutionContext{
			User: map[string]interface{}{
				// TODO: загрузить из БД
			},
			Execution: map[string]interface{}{
				"id": msg.ExecutionID,
			},
			Steps:     make(map[string]nodes.StepOutput),
			Variables: make(map[string]interface{}),
		},
		UpdatedAt: time.Now(),
	}
}

// loadExecutionState загружает состояние из БД
func (e *Engine) loadExecutionState(ctx context.Context, tx *sql.Tx, executionID string) (*ExecutionState, error) {
	var state ExecutionState
	var contextJSON []byte

	err := tx.QueryRowContext(ctx, `
		SELECT execution_id, current_node_id, context, updated_at
		FROM main.execution_state
		WHERE execution_id = $1
	`, executionID).Scan(&state.ExecutionID, &state.CurrentNodeID, &contextJSON, &state.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Первый запуск
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contextJSON, &state.Context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &state, nil
}

// loadSchema загружает схему из БД
func (e *Engine) loadSchema(ctx context.Context, tx *sql.Tx, schemaID int64) (*SchemaDefinition, error) {
	var defJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT definition FROM main.schemas WHERE id = $1
	`, schemaID).Scan(&defJSON)

	if err != nil {
		return nil, err
	}

	var schema SchemaDefinition
	if err := json.Unmarshal(defJSON, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema definition: %w", err)
	}

	return &schema, nil
}

// findNode находит ноду в схеме
func (e *Engine) findNode(schema *SchemaDefinition, nodeID string) *nodes.Node {
	for i := range schema.Nodes {
		if schema.Nodes[i].ID == nodeID {
			return &schema.Nodes[i]
		}
	}
	return nil
}

// saveExecutionStep сохраняет шаг в БД
func (e *Engine) saveExecutionStep(
	ctx context.Context,
	tx *sql.Tx,
	executionID string,
	node *nodes.Node,
	result *nodes.NodeResult,
	startedAt, finishedAt time.Time,
) error {
	outputJSON, err := json.Marshal(result.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	status := int16(1) // success
	if result.Status == nodes.StatusFailed {
		status = 2
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO main.execution_steps (
			execution_id, node_id, node_type,
			output, id_status, error,
			started_at, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, executionID, node.ID, node.Data.Type, outputJSON, status, result.Error, startedAt, finishedAt)

	return err
}

// updateContext обновляет контекст после выполнения ноды
func (e *Engine) updateContext(ctx *nodes.ExecutionContext, nodeID string, result *nodes.NodeResult) {
	ctx.Steps[nodeID] = nodes.StepOutput{
		Output: result.Output,
	}
}

// findNextNode находит следующую ноду через edges
func (e *Engine) findNextNode(schema *SchemaDefinition, currentNodeID string) *string {
	for _, edge := range schema.Edges {
		if edge.Source == currentNodeID {
			return &edge.Target
		}
	}
	return nil
}

// saveExecutionState сохраняет состояние в БД
func (e *Engine) saveExecutionState(ctx context.Context, tx *sql.Tx, state *ExecutionState) error {
	contextJSON, err := json.Marshal(state.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO main.execution_state (execution_id, current_node_id, context, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (execution_id) DO UPDATE SET
			current_node_id = EXCLUDED.current_node_id,
			context = EXCLUDED.context,
			updated_at = EXCLUDED.updated_at
	`, state.ExecutionID, state.CurrentNodeID, contextJSON, state.UpdatedAt)

	return err
}

// updateExecutionStatus обновляет статус execution
func (e *Engine) updateExecutionStatus(
	ctx context.Context,
	tx *sql.Tx,
	msg *ExecutionMessage,
	result *nodes.NodeResult,
	nextNodeID *string,
) error {
	// Определяем новый статус
	var newStatus int16
	var finishedAt *time.Time
	var errorMsg *string

	switch result.Status {
	case nodes.StatusFailed:
		newStatus = 5 // failed
		now := time.Now()
		finishedAt = &now
		errorMsg = result.Error

	case nodes.StatusSuccess:
		if nextNodeID == nil {
			// Это была последняя нода (End)
			newStatus = 4 // completed
			now := time.Now()
			finishedAt = &now
		} else {
			newStatus = 2 // running
		}

	case nodes.StatusSleep:
		newStatus = 3 // paused
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE main.executions
		SET id_status = $1, current_step_id = $2, finished_at = $3, error = $4
		WHERE id = $5
	`, newStatus, msg.CurrentNodeID, finishedAt, errorMsg, msg.ExecutionID)

	return err
}

// GetNextNodeID возвращает ID следующей ноды для публикации в RabbitMQ
func (e *Engine) GetNextNodeID(ctx context.Context, executionID string) (*string, error) {
	var state ExecutionState

	err := e.db.QueryRowContext(ctx, `
		SELECT current_node_id FROM main.execution_state WHERE execution_id = $1
	`, executionID).Scan(&state.CurrentNodeID)

	if err != nil {
		return nil, err
	}

	return &state.CurrentNodeID, nil
}

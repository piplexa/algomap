package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/piplexa/algomap/internal/domain"
	"go.uber.org/zap"
)

// ExecutionRepository предоставляет методы для работы с executions
type ExecutionRepository struct {
	db     *DB
	logger *zap.Logger
}

// NewExecutionRepository создаёт новый репозиторий executions
func NewExecutionRepository(db *DB, logger *zap.Logger) *ExecutionRepository {
	return &ExecutionRepository{
		db:     db,
		logger: logger,
	}
}

// Create создаёт новое выполнение схемы
func (r *ExecutionRepository) Create(ctx context.Context, req *domain.CreateExecutionRequest, createdBy int64) (*domain.Execution, error) {
	// Генерируем UUID
	executionID := uuid.New()

	// Сериализуем trigger_payload в JSON
	var payloadJSON []byte
	if req.TriggerPayload != nil {
		payloadJSON = req.TriggerPayload
	} else {
		payloadJSON = json.RawMessage("{}")
	}

	query := `
		INSERT INTO main.executions (
			id, schema_id, id_status, id_trigger_type, 
			trigger_payload, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, schema_id, id_status, id_trigger_type, trigger_payload, 
		          current_step_id, started_at, finished_at, created_at, created_by, error
	`

	var exec domain.Execution
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		executionID,
		req.SchemaID,
		1, // status: pending
		1, // trigger_type: manual
		payloadJSON,
		createdBy,
		time.Now().UTC(),
	).Scan(
		&exec.ID,
		&exec.SchemaID,
		&exec.StatusID,
		&exec.TriggerTypeID,
		&exec.TriggerPayload,
		&exec.CurrentStepID,
		&exec.StartedAt,
		&exec.FinishedAt,
		&exec.CreatedAt,
		&exec.CreatedBy,
		&exec.Error,
	)

	if err != nil {
		r.logger.Error("Failed to create execution",
			zap.Error(err),
			zap.Int64("schema_id", req.SchemaID),
		)
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	r.logger.Info("Execution created successfully",
		zap.String("execution_id", exec.ID),
		zap.Int64("schema_id", exec.SchemaID),
	)

	return &exec, nil
}

// GetByID получает execution по ID
func (r *ExecutionRepository) GetByID(ctx context.Context, id string) (*domain.Execution, error) {
	// Парсим UUID
	executionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	query := `
		SELECT 
			id, schema_id, id_status, id_trigger_type, trigger_payload,
			current_step_id, started_at, finished_at, created_at, created_by, error
		FROM main.executions
		WHERE id = $1
	`

	var exec domain.Execution
	err = r.db.Pool.QueryRow(ctx, query, executionID).Scan(
		&exec.ID,
		&exec.SchemaID,
		&exec.StatusID,
		&exec.TriggerTypeID,
		&exec.TriggerPayload,
		&exec.CurrentStepID,
		&exec.StartedAt,
		&exec.FinishedAt,
		&exec.CreatedAt,
		&exec.CreatedBy,
		&exec.Error,
	)

	if err != nil {
		r.logger.Error("Failed to get execution",
			zap.Error(err),
			zap.String("execution_id", id),
		)
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &exec, nil
}

// UpdateStatus обновляет статус execution
func (r *ExecutionRepository) UpdateStatus(ctx context.Context, id string, status int16, errorMsg string) error {
	executionID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid execution ID: %w", err)
	}

	query := `
		UPDATE main.executions
		SET 
			id_status = $2,
			error = $3,
			finished_at = CASE 
				WHEN $2 IN (4, 5, 6) THEN NOW() 
				ELSE finished_at 
			END
		WHERE id = $1
	`

	var errPtr *string
	if errorMsg != "" {
		errPtr = &errorMsg
	}

	result, err := r.db.Pool.Exec(ctx, query, executionID, status, errPtr)
	if err != nil {
		r.logger.Error("Failed to update execution status",
			zap.Error(err),
			zap.String("execution_id", id),
			zap.Int16("status", status),
		)
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("execution with id %s not found", id)
	}

	r.logger.Info("Execution status updated",
		zap.String("execution_id", id),
		zap.Int16("status", status),
	)

	return nil
}

// GetSteps возвращает историю шагов выполнения
func (r *ExecutionRepository) GetSteps(ctx context.Context, id string) ([]*domain.ExecutionStep, error) {
	executionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	query := `
		SELECT 
			id, execution_id, node_id, node_type, prev_node_id, next_node_id,
			input, output, id_status, error, started_at, finished_at
		FROM main.execution_steps
		WHERE execution_id = $1
		ORDER BY started_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, executionID)
	if err != nil {
		r.logger.Error("Failed to query execution steps",
			zap.Error(err),
			zap.String("execution_id", id),
		)
		return nil, fmt.Errorf("failed to query steps: %w", err)
	}
	defer rows.Close()

	var steps []*domain.ExecutionStep
	for rows.Next() {
		var step domain.ExecutionStep
		err := rows.Scan(
			&step.ID,
			&step.ExecutionID,
			&step.NodeID,
			&step.NodeType,
			&step.PrevNodeID,
			&step.NextNodeID,
			&step.Input,
			&step.Output,
			&step.StatusID,
			&step.Error,
			&step.StartedAt,
			&step.FinishedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan execution step",
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		steps = append(steps, &step)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating execution steps", zap.Error(err))
		return nil, fmt.Errorf("error iterating steps: %w", err)
	}

	return steps, nil
}

// GetState возвращает текущее состояние выполнения
func (r *ExecutionRepository) GetState(ctx context.Context, id string) (*domain.ExecutionState, error) {
	executionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	query := `
		SELECT execution_id, current_node_id, context, updated_at
		FROM main.execution_state
		WHERE execution_id = $1
	`

	var state domain.ExecutionState
	err = r.db.Pool.QueryRow(ctx, query, executionID).Scan(
		&state.ExecutionID,
		&state.CurrentNodeID,
		&state.Context,
		&state.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get execution state",
			zap.Error(err),
			zap.String("execution_id", id),
		)
		return nil, fmt.Errorf("failed to get execution state: %w", err)
	}

	return &state, nil
}

// CreateOrUpdateState создаёт или обновляет состояние выполнения
func (r *ExecutionRepository) CreateOrUpdateState(ctx context.Context, state *domain.ExecutionState) error {
	query := `
		INSERT INTO main.execution_state (execution_id, current_node_id, context, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (execution_id) 
		DO UPDATE SET 
			current_node_id = EXCLUDED.current_node_id,
			context = EXCLUDED.context,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		state.ExecutionID,
		state.CurrentNodeID,
		state.Context,
		time.Now().UTC(),
	)

	if err != nil {
		r.logger.Error("Failed to create/update execution state",
			zap.Error(err),
			zap.String("execution_id", state.ExecutionID),
		)
		return fmt.Errorf("failed to create/update state: %w", err)
	}

	return nil
}

// CreateStep создаёт новый шаг выполнения
func (r *ExecutionRepository) CreateStep(ctx context.Context, step *domain.ExecutionStep) error {
	query := `
		INSERT INTO main.execution_steps (
			execution_id, node_id, node_type, prev_node_id, next_node_id,
			input, output, id_status, error, started_at, finished_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		step.ExecutionID,
		step.NodeID,
		step.NodeType,
		step.PrevNodeID,
		step.NextNodeID,
		step.Input,
		step.Output,
		step.StatusID,
		step.Error,
		step.StartedAt,
		step.FinishedAt,
	).Scan(&step.ID)

	if err != nil {
		r.logger.Error("Failed to create execution step",
			zap.Error(err),
			zap.String("execution_id", step.ExecutionID),
			zap.String("node_id", step.NodeID),
		)
		return fmt.Errorf("failed to create step: %w", err)
	}

	return nil
}


// ExecutionRepository - репозиторий для работы с выполнениями схем
// TODO: Реализовать методы:
// - Create() - создать выполнение
// - GetByID() - получить выполнение по ID
// - List() - список выполнений (с фильтрацией по schema_id, status)
// - UpdateStatus() - обновить статус выполнения
// - SaveState() - сохранить состояние выполнения
// - LoadState() - загрузить состояние выполнения
// - CreateStep() - создать шаг выполнения
// - GetSteps() - получить все шаги выполнения
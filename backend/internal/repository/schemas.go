package repository

import (
	"context"
	"fmt"

	"github.com/piplexa/algomap/internal/domain"
	"go.uber.org/zap"
)

// SchemaRepository предоставляет методы для работы со схемами
type SchemaRepository struct {
	db     *DB
	logger *zap.Logger
}

// NewSchemaRepository создаёт новый репозиторий схем
func NewSchemaRepository(db *DB, logger *zap.Logger) *SchemaRepository {
	return &SchemaRepository{
		db:     db,
		logger: logger,
	}
}

// Create создаёт новую схему
func (r *SchemaRepository) Create(ctx context.Context, req *domain.CreateSchemaRequest, createdBy int64) (*domain.Schema, error) {
	query := `
		INSERT INTO main.schemas (name, description, definition, id_status, created_by)
		VALUES ($1, $2, $3, 2, $4)
		RETURNING id, name, description, definition, id_status, created_by, created_at, updated_at
	`

	var schema domain.Schema
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		req.Name,
		req.Description,
		req.Definition,
		createdBy,
	).Scan(
		&schema.ID,
		&schema.Name,
		&schema.Description,
		&schema.Definition,
		&schema.Status,
		&schema.CreatedBy,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create schema",
			zap.Error(err),
			zap.String("name", req.Name),
		)
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	r.logger.Info("Schema created successfully",
		zap.Int64("schema_id", schema.ID),
		zap.String("name", schema.Name),
	)

	return &schema, nil
}

// GetByID получает схему по ID
func (r *SchemaRepository) GetByID(ctx context.Context, id int64) (*domain.Schema, error) {
	query := `
		SELECT id, name, description, definition, id_status, created_by, created_at, updated_at
		FROM main.schemas
		WHERE id = $1
	`

	var schema domain.Schema
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&schema.ID,
		&schema.Name,
		&schema.Description,
		&schema.Definition,
		&schema.Status,
		&schema.CreatedBy,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get schema",
			zap.Error(err),
			zap.Int64("schema_id", id),
		)
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	return &schema, nil
}

// List возвращает список схем с опциональной фильтрацией
func (r *SchemaRepository) List(ctx context.Context, status *int16, limit, offset int, id_user int64) ([]*domain.Schema, error) {
	query := `
		SELECT id, name, description, definition, id_status, created_by, created_at, updated_at
		FROM main.schemas
		WHERE ($1::SMALLINT IS NULL OR id_status = $1) and created_by = $4
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, status, limit, offset, id_user)
	if err != nil {
		r.logger.Error("Failed to list schemas", zap.Error(err))
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*domain.Schema
	for rows.Next() {
		var schema domain.Schema
		err := rows.Scan(
			&schema.ID,
			&schema.Name,
			&schema.Description,
			&schema.Definition,
			&schema.Status,
			&schema.CreatedBy,
			&schema.CreatedAt,
			&schema.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan schema", zap.Error(err))
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, &schema)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating schemas", zap.Error(err))
		return nil, fmt.Errorf("error iterating schemas: %w", err)
	}

	return schemas, nil
}

// Update обновляет схему
func (r *SchemaRepository) Update(ctx context.Context, id int64, req *domain.UpdateSchemaRequest) (*domain.Schema, error) {
	// Динамически строим запрос в зависимости от того, что нужно обновить
	query := `
		UPDATE main.schemas
		SET 
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			definition = COALESCE($4, definition),
			id_status = COALESCE($5, id_status),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, definition, id_status, created_by, created_at, updated_at
	`

	var schema domain.Schema
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		id,
		req.Name,
		req.Description,
		req.Definition,
		req.Status,
	).Scan(
		&schema.ID,
		&schema.Name,
		&schema.Description,
		&schema.Definition,
		&schema.Status,
		&schema.CreatedBy,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update schema",
			zap.Error(err),
			zap.Int64("schema_id", id),
		)
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}

	r.logger.Info("Schema updated successfully",
		zap.Int64("schema_id", schema.ID),
		zap.String("name", schema.Name),
	)

	return &schema, nil
}

// Delete удаляет схему
func (r *SchemaRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM main.schemas WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete schema",
			zap.Error(err),
			zap.Int64("schema_id", id),
		)
		return fmt.Errorf("failed to delete schema: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("schema with id %d not found", id)
	}

	r.logger.Info("Schema deleted successfully", zap.Int64("schema_id", id))

	return nil
}
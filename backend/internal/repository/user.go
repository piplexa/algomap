package repository

import (
	"context"
	"fmt"

	"github.com/piplexa/algomap/internal/domain"
	"go.uber.org/zap"
)

// UserRepository предоставляет методы для работы с пользователями
type UserRepository struct {
	db     *DB
	logger *zap.Logger
}

// NewUserRepository создаёт новый репозиторий пользователей
func NewUserRepository(db *DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create создаёт нового пользователя
func (r *UserRepository) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	query := `
		INSERT INTO main.users (email, name, hashPassword)
		VALUES ($1, $2, crypt($3, gen_salt('bf')))
		RETURNING id, email, name, created_at
	`

	var user domain.User
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		req.Email,
		req.Name,
		req.Password,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return &user, nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, name, created_at
		FROM main.users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get user",
			zap.Error(err),
			zap.Int64("user_id", id),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, created_at
		FROM main.users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get user by email",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// List возвращает список пользователей
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, email, name, created_at
		FROM main.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan user", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating users", zap.Error(err))
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Update обновляет пользователя
func (r *UserRepository) Update(ctx context.Context, id int64, req *domain.UpdateUserRequest) (*domain.User, error) {
	query := `
		UPDATE main.users
		SET 
			name = COALESCE($2, name)
		WHERE id = $1
		RETURNING id, email, name, created_at
	`

	var user domain.User
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		id,
		req.Name,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to update user",
			zap.Error(err),
			zap.Int64("user_id", id),
		)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	r.logger.Info("User updated successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return &user, nil
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM main.users WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user",
			zap.Error(err),
			zap.Int64("user_id", id),
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	r.logger.Info("User deleted successfully", zap.Int64("user_id", id))

	return nil
}

// VerifyPassword проверяет соответствие пароля хешу
func (r *UserRepository) VerifyPassword(ctx context.Context, email, password string) (*domain.User, error) {
	query := `
		SELECT id, email, name, created_at
		FROM main.users
		WHERE email = $1
		  AND hashPassword = crypt($2, hashPassword)
	`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email, password).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to verify password",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	return &user, nil
}
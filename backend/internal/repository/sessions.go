package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/piplexa/algomap/internal/domain"
	"go.uber.org/zap"
)

// SessionRepository предоставляет методы для работы с сессиями
type SessionRepository struct {
	db     *DB
	logger *zap.Logger
}

// NewSessionRepository создаёт новый репозиторий сессий
func NewSessionRepository(db *DB, logger *zap.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create создаёт новую сессию
func (r *SessionRepository) Create(ctx context.Context, userID int64, ttl time.Duration) (*domain.Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO main.sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, expires_at, created_at
	`

	var session domain.Session
	err := r.db.Pool.QueryRow(
		ctx,
		query,
		sessionID,
		userID,
		expiresAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Info("Session created successfully",
		zap.String("session_id", session.ID),
		zap.Int64("user_id", userID),
	)

	return &session, nil
}

// GetByID получает сессию по ID (session_key)
func (r *SessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM main.sessions
		WHERE id = $1 AND expires_at > NOW()
	`

	var session domain.Session
	err := r.db.Pool.QueryRow(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		// Не логируем как ошибку, т.к. это может быть просто невалидная сессия
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	return &session, nil
}

// Delete удаляет сессию (logout)
func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM main.sessions WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, sessionID)
	if err != nil {
		r.logger.Error("Failed to delete session",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	r.logger.Info("Session deleted successfully", zap.String("session_id", sessionID))

	return nil
}

// DeleteAllByUserID удаляет все сессии пользователя
func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM main.sessions WHERE user_id = $1`

	result, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user sessions",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	r.logger.Info("User sessions deleted",
		zap.Int64("user_id", userID),
		zap.Int64("count", result.RowsAffected()),
	)

	return nil
}

// CleanupExpired удаляет истёкшие сессии (можно запускать по крону)
func (r *SessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM main.sessions WHERE expires_at < NOW()`

	result, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		r.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	count := result.RowsAffected()
	if count > 0 {
		r.logger.Info("Expired sessions cleaned up", zap.Int64("count", count))
	}

	return count, nil
}
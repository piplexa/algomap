package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DB хранит пул соединений с PostgreSQL
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB создаёт новое подключение к БД
func NewDB(databaseURL string, logger *zap.Logger) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Настройки пула
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL",
		zap.Int32("max_conns", config.MaxConns),
		zap.Int32("min_conns", config.MinConns),
	)

	return &DB{Pool: pool}, nil
}

// Close закрывает все соединения с БД
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping проверяет соединение с БД
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
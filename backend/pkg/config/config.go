package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config содержит все настройки приложения
type Config struct {
	// HTTP server
	Port string

	// Database
	DatabaseURL string

	// RabbitMQ
	RabbitMQURL string

	// Logging
	LogLevel string

	// AT Scheduler
	ATSchedulerURL string
	URLExecution   string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Пытаемся загрузить .env файл (опционально)
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RabbitMQURL:    getEnv("RABBITMQ_URL", ""),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		ATSchedulerURL: getEnv("AT_SCHEDULER_URL", ""),
		URLExecution:   getEnv("URL_EXECUTION", ""),
	}

	// Валидация обязательных параметров
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// getEnv читает переменную окружения или возвращает defaultValue
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
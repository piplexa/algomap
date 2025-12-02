package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/piplexa/algomap/internal/executor"
	"github.com/piplexa/algomap/internal/nodes"
	"github.com/piplexa/algomap/pkg/rabbitmq"

	"github.com/piplexa/algomap/pkg/config"
	"github.com/piplexa/algomap/pkg/logger"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Инициализируем логгер
	if err := logger.Init(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting AlgoMap Worker",
		zap.String("log_level", cfg.LogLevel),
	)

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	logger.Info("connected to database")

	// Создаём реестр обработчиков нод
	registry := nodes.NewHandlerRegistry()
	registry.Register("start", nodes.NewStartHandler())
	registry.Register("end", nodes.NewEndHandler())
	registry.Register("log", nodes.NewLogHandler(logger.Log))
	registry.Register("variable_set", nodes.NewVariableSetHandler())
	registry.Register("math", nodes.NewMathHandler())
	registry.Register("sleep", nodes.NewSleepHandler(cfg.ATSchedulerURL, cfg.URLExecution, logger.Log))
	registry.Register("http_request", nodes.NewHTTPRequestHandler())
	registry.Register("condition", nodes.NewConditionHandler())
	// TODO: Добавить остальные обработчики

	// Создаём движок выполнения
	engine := executor.NewEngine(db, logger.Log, registry)

	// Создаём consumer
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL, logger.Log, engine)
	if err != nil {
		logger.Fatal("failed to create consumer", zap.Error(err))
	}
	defer consumer.Close()

	// Создаём контекст с обработкой сигналов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
	}()

	// Запускаем worker
	if err := consumer.Start(ctx); err != nil {
		logger.Error("worker stopped with error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("worker stopped gracefully")
}

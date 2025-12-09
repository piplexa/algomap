package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/piplexa/algomap/internal/handlers"
	authmiddleware "github.com/piplexa/algomap/internal/middleware"
	"github.com/piplexa/algomap/internal/repository"
	"github.com/piplexa/algomap/pkg/config"
	"github.com/piplexa/algomap/pkg/logger"
	"github.com/piplexa/algomap/pkg/rabbitmq"
	"go.uber.org/zap"

	_ "github.com/piplexa/algomap/docs" // swagger docs
    httpSwagger "github.com/swaggo/http-swagger"
)

// @title           AlgoMap API
// @version         1.0
// @description     Visual workflow automation platform API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/piplexa/algomap
// @contact.email  piplexa@list.ru

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session_id
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

	logger.Info("Starting AlgoMap API server",
		zap.String("port", cfg.Port),
		zap.String("log_level", cfg.LogLevel),
	)

	// 3. Подключаемся к PostgreSQL
	db, err := repository.NewDB(cfg.DatabaseURL, logger.Log)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// 4. Подключаемся к RabbitMQ
	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQURL, logger.Log)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rmqConn.Close()

	// Создаём publisher
	rmqPublisher := rabbitmq.NewPublisher(rmqConn, logger.Log)

	// Объявляем очередь для выполнения схем
	queueName := "schema_execution_queue"
	if err := rmqPublisher.DeclareQueue(queueName); err != nil {
		logger.Fatal("Failed to declare queue", zap.Error(err), zap.String("queue", queueName))
	}

	// 5. Создаём репозитории
	userRepo := repository.NewUserRepository(db, logger.Log)
	sessionRepo := repository.NewSessionRepository(db, logger.Log)
	schemaRepo := repository.NewSchemaRepository(db, logger.Log)
	executionRepo := repository.NewExecutionRepository(db, logger.Log)

	// 6. Создаём handlers
	userHandler := handlers.NewUserHandler(userRepo, logger.Log)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, logger.Log)
	schemaHandler := handlers.NewSchemaHandler(schemaRepo, logger.Log)
	executionHandler := handlers.NewExecutionHandler(executionRepo, schemaRepo, logger.Log, rmqPublisher, queueName)

	// 7. Создаём middleware
	authMw := authmiddleware.NewAuthMiddleware(sessionRepo, logger.Log)

	// 8. Настраиваем роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS (для локальной разработки)
	r.Use(corsMiddleware)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Публичные endpoints (без auth)
		r.Post("/users/register", userHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		//
		// TODO: Вот тут надо подумать, как правильно сделать
		// Скорее всего при вызове из вне, например из AT, нужно указывать какой-то API ключ, чтобы при получении 
		// запроса, тут было бы понятно, что это доверенный источник и можно запускать схему
		// Заметки. см. тут; nodes/sleep.go:105
		//
		r.Post("/executions/{id-execution}/{id-node}/continue", executionHandler.Continue)

		// Защищённые endpoints (требуют auth)
		r.Group(func(r chi.Router) {
			r.Use(authMw.RequireAuth)

			// Auth
			r.Post("/auth/logout", authHandler.Logout)

			// Пользователи
			r.Get("/users", userHandler.List)
			r.Get("/users/{id}", userHandler.GetByID)
			r.Put("/users/{id}", userHandler.Update)
			r.Delete("/users/{id}", userHandler.Delete)

			// Схемы
			r.Get("/schemas", schemaHandler.List)
			r.Post("/schemas", schemaHandler.Create)
			r.Get("/schemas/{id}", schemaHandler.GetByID)
			r.Put("/schemas/{id}", schemaHandler.Update)
			r.Delete("/schemas/{id}", schemaHandler.Delete)

			// Executions
			r.Post("/executions", executionHandler.Create)
			r.Get("/executions/{id}", executionHandler.GetByID)
			r.Get("/executions/{id}/steps", executionHandler.GetSteps)
			r.Get("/executions/{id}/state", executionHandler.GetState)
			r.Get("/executions/list/{id}", executionHandler.GetExecutionsBySchemaID)
			r.Delete("/executions/schema/{id}", executionHandler.DeleteBySchemaID)

			// TODO: Webhook endpoints
		})
	})

	// 9. Запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Запускаем в горутине
	go func() {
		logger.Info("API server listening", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 10. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// corsMiddleware добавляет CORS заголовки (для локальной разработки)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

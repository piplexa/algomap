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
	"go.uber.org/zap"
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

	// 4. Создаём репозитории
	userRepo := repository.NewUserRepository(db, logger.Log)
	sessionRepo := repository.NewSessionRepository(db, logger.Log)
	schemaRepo := repository.NewSchemaRepository(db, logger.Log)

	// 5. Создаём handlers
	userHandler := handlers.NewUserHandler(userRepo, logger.Log)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, logger.Log)
	schemaHandler := handlers.NewSchemaHandler(schemaRepo, logger.Log)

	// 6. Создаём middleware
	authMw := authmiddleware.NewAuthMiddleware(sessionRepo, logger.Log)

	// 7. Настраиваем роутер
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

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Публичные endpoints (без auth)
		r.Post("/users/register", userHandler.Register)
		r.Post("/auth/login", authHandler.Login)

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

			// TODO: Executions endpoints
			// TODO: Webhook endpoints
		})
	})

	// 8. Запускаем HTTP сервер
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		logger.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// corsMiddleware добавляет CORS headers (для локальной разработки)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
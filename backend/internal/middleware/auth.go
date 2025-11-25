package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/piplexa/algomap/internal/repository"
	"go.uber.org/zap"
)

// ContextKey тип для ключей в context
type ContextKey string

const (
	// UserIDKey ключ для user_id в context
	UserIDKey ContextKey = "user_id"
)

// AuthMiddleware middleware для проверки аутентификации
type AuthMiddleware struct {
	sessionRepo *repository.SessionRepository
	logger      *zap.Logger
}

// NewAuthMiddleware создаёт новый auth middleware
func NewAuthMiddleware(sessionRepo *repository.SessionRepository, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// RequireAuth проверяет наличие валидной сессии
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем session_key из Cookie или Bearer token
		sessionID := m.getSessionID(r)
		if sessionID == "" {
			m.respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Проверяем сессию в БД
		session, err := m.sessionRepo.GetByID(r.Context(), sessionID)
		if err != nil {
			m.respondError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		// Добавляем user_id в context
		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)

		// Передаём запрос дальше
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getSessionID извлекает session_key из Cookie или Authorization header
func (m *AuthMiddleware) getSessionID(r *http.Request) string {
	// Сначала пытаемся получить из Cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}

	// Если нет в Cookie, пытаемся получить из Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Формат: "Bearer <session_key>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}

// respondError отправляет JSON ответ с ошибкой
func (m *AuthMiddleware) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserIDFromContext извлекает user_id из context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
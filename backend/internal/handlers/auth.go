package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/piplexa/algomap/internal/domain"
	"github.com/piplexa/algomap/internal/repository"
	"go.uber.org/zap"
)

// AuthHandler обрабатывает запросы для аутентификации
type AuthHandler struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	logger      *zap.Logger
	sessionTTL  time.Duration
}

// NewAuthHandler создаёт новый handler для аутентификации
func NewAuthHandler(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
		sessionTTL:  24 * time.Hour * 7, // 7 дней по умолчанию
	}
}

// Login аутентифицирует пользователя и создаёт сессию
// @Summary      Вход в систему
// @Description  Аутентификация пользователя по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Email и пароль"
// @Success      200          {object}  LoginResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		h.respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Проверяем пароль и получаем пользователя
	user, err := h.userRepo.VerifyPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// TODO: OAuth провайдеры (Google, GitHub)
	// Здесь будет логика для OAuth:
	// 1. Редирект на провайдера
	// 2. Получение токена
	// 3. Получение email от провайдера
	// 4. Создание/поиск пользователя
	// 5. Создание сессии

	// Создаём сессию
	session, err := h.sessionRepo.Create(r.Context(), user.ID, h.sessionTTL)
	if err != nil {
		h.logger.Error("Failed to create session", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Возвращаем session_key и информацию о пользователе
	response := domain.LoginResponse{
		SessionKey: session.ID,
		User:       user,
		ExpiresAt:  session.ExpiresAt,
	}

	h.logger.Info("User logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
	)

	h.respondJSON(w, http.StatusOK, response)
}

// Logout завершает сессию пользователя
// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем session_key из Cookie или Bearer token
	sessionID := h.getSessionID(r)
	if sessionID == "" {
		h.respondError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Удаляем сессию
	if err := h.sessionRepo.Delete(r.Context(), sessionID); err != nil {
		h.logger.Error("Failed to delete session", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	h.logger.Info("User logged out successfully", zap.String("session_id", sessionID))

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// getSessionID извлекает session_key из Cookie или Authorization header
func (h *AuthHandler) getSessionID(r *http.Request) string {
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

// respondJSON отправляет JSON ответ
func (h *AuthHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError отправляет JSON ответ с ошибкой
func (h *AuthHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/piplexa/algomap/internal/domain"
	"github.com/piplexa/algomap/internal/repository"
	"go.uber.org/zap"
)

// UserHandler обрабатывает запросы для пользователей
type UserHandler struct {
	repo   *repository.UserRepository
	logger *zap.Logger
}

// NewUserHandler создаёт новый handler для пользователей
func NewUserHandler(repo *repository.UserRepository, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		repo:   repo,
		logger: logger,
	}
}

// Register регистрирует нового пользователя
// POST /api/users/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация Email
	if req.Email == "" {
		h.respondError(w, http.StatusBadRequest, "Email is required")
		return
	}
	// Валидация password
	if req.Password == "" {
		h.respondError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Проверяем, что пользователь с таким email ещё не существует
	existingUser, _ := h.repo.GetByEmail(r.Context(), req.Email)
	if existingUser != nil {
		h.respondError(w, http.StatusConflict, "User with this email already exists")
		return
	}

	user, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}

// GetByID возвращает пользователя по ID
// GET /api/users/:id
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found")
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// List возвращает список пользователей
// GET /api/users?limit=10&offset=0
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	h.respondJSON(w, http.StatusOK, users)
}

// Update обновляет пользователя
// PUT /api/users/:id
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	h.respondJSON(w, http.StatusOK, user)
}

// Delete удаляет пользователя
// DELETE /api/users/:id
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

// respondJSON отправляет JSON ответ
func (h *UserHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError отправляет JSON ответ с ошибкой
func (h *UserHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
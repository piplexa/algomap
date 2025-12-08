package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/piplexa/algomap/internal/domain"
	"github.com/piplexa/algomap/internal/middleware"
	"github.com/piplexa/algomap/internal/repository"
	"go.uber.org/zap"

	"reflect"
)

// SchemaHandler обрабатывает запросы для схем
type SchemaHandler struct {
	repo   *repository.SchemaRepository
	logger *zap.Logger
}

// NewSchemaHandler создаёт новый handler для схем
func NewSchemaHandler(repo *repository.SchemaRepository, logger *zap.Logger) *SchemaHandler {
	return &SchemaHandler{
		repo:   repo,
		logger: logger,
	}
}

// Create создаёт новую схему
// POST /api/schemas
func (h *SchemaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Получаем user_id из context (установлен в auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	schema, err := h.repo.Create(r.Context(), &req, userID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create schema")
		return
	}

	h.respondJSON(w, http.StatusCreated, schema)
}

// GetByID возвращает схему по ID
// GET /api/schemas/:id
func (h *SchemaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid schema ID")
		return
	}

	schema, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Schema not found")
		return
	}

	h.respondJSON(w, http.StatusOK, schema)
}

// List возвращает список схем
// GET /api/schemas?status=1&limit=10&offset=0
func (h *SchemaHandler) List(w http.ResponseWriter, r *http.Request) {
	// Парсим query параметры
	var status *int16
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		s, err := strconv.ParseInt(statusStr, 10, 16)
		if err == nil {
			statusVal := int16(s)
			status = &statusVal
		}
	}

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

	// Получаем id пользователя из context
	id_user, _ := r.Context().Value(middleware.UserIDKey).(int64)

	schemas, err := h.repo.List(r.Context(), status, limit, offset, id_user)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list schemas")
		return
	}

	h.respondJSON(w, http.StatusOK, schemas)
}

// Update обновляет схему
// PUT /api/schemas/:id
func (h *SchemaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid schema ID")
		return
	}

	var req domain.UpdateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	schema, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to update schema")
		return
	}

	h.respondJSON(w, http.StatusOK, schema)
}

// Delete удаляет схему
// DELETE /api/schemas/:id
func (h *SchemaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid schema ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to delete schema")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Schema deleted successfully",
	})
}

// respondJSON отправляет JSON ответ
func (h *SchemaHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if isNilValue(data) {
		h.logger.Warn("respondJSON called with nil data - this should be fixed in repository!")
		// TODO: Подумать: пусть репозитории возвращают пустой слайс или мапу вместо nil или делать это тут?
    
		value := reflect.ValueOf(data)
		if value.Kind() == reflect.Slice {
			data = []interface{}{}
		} else {
			data = map[string]interface{}{}
		}
    }

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError отправляет JSON ответ с ошибкой
func (h *SchemaHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
package handlers

// ExecutionHandler - HTTP handlers для управления выполнениями схем

// Реализованные endpoints:
// POST   /api/executions                	- запустить схему (manual)
// GET    /api/executions/:id/steps      	- история шагов
// GET    /api/executions/:id            	- статус выполнения
// POST   /api/executions/:id-execution/:id-node/continue  - начать выполнение с указанного узла схемы
// GET	  /api/executions/list/:id-schema	- список выполнений с фильтрацией по схеме

// TODO: Реализовать endpoints:
// POST   /api/executions/:id/pause      	- пауза
// POST   /api/executions/:id/resume     	- продолжить
// POST   /api/executions/:id/stop       	- остановить
// GET    /api/executions/:id/logs       	- логи выполнения
// POST   /api/executions/:id/:id/one    	- выполнить только указанный узел схемы

import (
	"context"
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

// RabbitMQPublisher интерфейс для публикации сообщений
type RabbitMQPublisher interface {
	Publish(ctx context.Context, queueName string, message interface{}) error
}

// ExecutionHandler обрабатывает запросы для executions
type ExecutionHandler struct {
	execRepo     *repository.ExecutionRepository
	schemaRepo   *repository.SchemaRepository
	logger       *zap.Logger
	rmqPublisher RabbitMQPublisher
	queueName    string
}

// NewExecutionHandler создаёт новый handler для executions
func NewExecutionHandler(
	execRepo *repository.ExecutionRepository,
	schemaRepo *repository.SchemaRepository,
	logger *zap.Logger,
	rmqPublisher RabbitMQPublisher,
	queueName string,
) *ExecutionHandler {
	return &ExecutionHandler{
		execRepo:     execRepo,
		schemaRepo:   schemaRepo,
		logger:       logger,
		rmqPublisher: rmqPublisher,
		queueName:    queueName,
	}
}

// GetExecutionsBySchemaID возвращает список выполнений по ID схемы
// GET /api/executions/list/:id
func (h *ExecutionHandler) GetExecutionsBySchemaID(w http.ResponseWriter, r *http.Request) {
	schemaIDStr := chi.URLParam(r, "id") // id схемы из URL
	schemaID, err := strconv.ParseInt(schemaIDStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid schema ID")
		return
	}

	// Получаем user_id из context (установлен в auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Получаем список выполнений по schemaID и id_user
	executions, err := h.execRepo.List(r.Context(), schemaID, userID, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list schemas")
		return
	}

	h.respondJSON(w, http.StatusOK, executions)
}

// Create создаёт новое выполнение схемы и отправляет задачу в RabbitMQ
// POST /api/executions
func (h *ExecutionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateExecutionRequest
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

	// Проверяем существование схемы
	schema, err := h.schemaRepo.GetByID(r.Context(), req.SchemaID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Schema not found")
		return
	}

	// Ищем стартовую ноду в схеме
	startNodeID, err := h.findStartNode(schema.Definition)
	if err != nil {
		h.logger.Error("Failed to find start node",
			zap.Error(err),
			zap.Int64("schema_id", req.SchemaID),
		)
		h.respondError(w, http.StatusBadRequest, "Schema must have a start node")
		return
	}

	// Проверяем что схема активна
	if schema.Status != domain.SchemaStatusActive {
		h.respondError(w, http.StatusBadRequest, "Schema is not active")
		return
	}

	// Создаём execution в БД
	execution, err := h.execRepo.Create(r.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Failed to create execution",
			zap.Error(err),
			zap.Int64("schema_id", req.SchemaID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to create execution")
		return
	}

	// Публикуем в RabbitMQ
	message := map[string]interface{}{
		"execution_id":    execution.ID,
		"schema_id":       execution.SchemaID,
		"current_node_id": startNodeID,
		"debug_mode":      req.DebugMode,
	}

	h.logger.Info("Publishing execution to RabbitMQ",
		zap.String("queueName", h.queueName),
		zap.Int64("schema_id", execution.SchemaID),
	)

	if err := h.rmqPublisher.Publish(r.Context(), h.queueName, message); err != nil {
		h.execRepo.UpdateStatus(r.Context(), execution.ID, 5, "Failed to publish to RabbitMQ")
		h.logger.Error("Failed to publish to RabbitMQ",
			zap.Error(err),
			zap.String("execution_id", execution.ID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to queue execution")
		return
	}

	h.logger.Info("Execution created successfully",
		zap.String("execution_id", execution.ID),
		zap.Int64("schema_id", execution.SchemaID),
	)

	h.respondJSON(w, http.StatusCreated, execution)
}

// GetByID возвращает execution по ID
// GET /api/executions/:id
func (h *ExecutionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id")

	execution, err := h.execRepo.GetByID(r.Context(), executionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Execution not found")
		return
	}

	h.respondJSON(w, http.StatusOK, execution)
}

// GetSteps возвращает историю шагов выполнения
// GET /api/executions/:id/steps
func (h *ExecutionHandler) GetSteps(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id")

	steps, err := h.execRepo.GetSteps(r.Context(), executionID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to fetch steps")
		return
	}

	h.respondJSON(w, http.StatusOK, steps)
}

// GetState возвращает текущее состояние выполнения
// GET /api/executions/:id/state
func (h *ExecutionHandler) GetState(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id")

	state, err := h.execRepo.GetState(r.Context(), executionID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Execution state not found")
		return
	}

	h.respondJSON(w, http.StatusOK, state)
}

// DeleteBySchemaID удаляет всю историю выполнений для указанной схемы
// DELETE /api/executions/schema/:id
func (h *ExecutionHandler) DeleteBySchemaID(w http.ResponseWriter, r *http.Request) {
	schemaIDStr := chi.URLParam(r, "id")
	schemaID, err := strconv.ParseInt(schemaIDStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid schema ID")
		return
	}

	// Получаем user_id из context (установлен в auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Удаляем все выполнения схемы
	err = h.execRepo.DeleteBySchemaID(r.Context(), schemaID, userID)
	if err != nil {
		h.logger.Error("Failed to delete executions by schema ID",
			zap.Error(err),
			zap.Int64("schema_id", schemaID),
			zap.Int64("user_id", userID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to delete executions")
		return
	}

	h.logger.Info("Executions deleted successfully",
		zap.Int64("schema_id", schemaID),
		zap.Int64("user_id", userID),
	)

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "All executions for the schema have been deleted successfully",
	})
}

// Continue продолжает выполнение схемы с указанного узла
// POST /api/executions/:id-execution/:id-node/continue
func (h *ExecutionHandler) Continue(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "id-execution")
	nodeID := chi.URLParam(r, "id-node")

	h.logger.Info("Continue execution request",
		zap.String("execution_id", executionID),
		zap.String("node_id", nodeID),
	)

	// Получаем execution из БД чтобы получить schema_id
	execution, err := h.execRepo.GetByID(r.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get execution",
			zap.Error(err),
			zap.String("execution_id", executionID),
		)
		h.respondError(w, http.StatusNotFound, "Execution not found")
		return
	}

	// Публикуем сообщение в RabbitMQ
	message := map[string]interface{}{
		"execution_id":    executionID,
		"schema_id":       execution.SchemaID,
		"current_node_id": nodeID,
		"debug_mode":      false,
	}

	h.logger.Info("Publishing continue execution to RabbitMQ",
		zap.String("queueName", h.queueName),
		zap.String("execution_id", executionID),
		zap.Int64("schema_id", execution.SchemaID),
		zap.String("node_id", nodeID),
	)

	if err := h.rmqPublisher.Publish(r.Context(), h.queueName, message); err != nil {
		h.logger.Error("Failed to publish continue message to RabbitMQ",
			zap.Error(err),
			zap.String("execution_id", executionID),
			zap.String("node_id", nodeID),
		)
		h.respondError(w, http.StatusInternalServerError, "Failed to queue execution")
		return
	}

	h.logger.Info("Continue execution published successfully",
		zap.String("execution_id", executionID),
		zap.String("node_id", nodeID),
		zap.Int64("schema_id", execution.SchemaID),
	)

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message":      "Execution continue queued successfully",
		"execution_id": executionID,
		"node_id":      nodeID,
	})
}

// respondJSON отправляет JSON ответ
func (h *ExecutionHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	//h.logger.Debug("Responding with JSON: ", zap.Any("data", data))
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
func (h *ExecutionHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

// ReactFlowNode представляет структуру ноды из ReactFlow
type ReactFlowNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ReactFlowDefinition представляет структуру схемы из ReactFlow
type ReactFlowDefinition struct {
	Nodes []ReactFlowNode `json:"nodes"`
}

// findStartNode ищет ноду с типом "start" в definition схемы
func (h *ExecutionHandler) findStartNode(definition json.RawMessage) (string, error) {
	var def ReactFlowDefinition
	if err := json.Unmarshal(definition, &def); err != nil {
		return "", err
	}

	for _, node := range def.Nodes {
		if node.Type == "start" {
			return node.ID, nil
		}
	}

	return "", &ErrStartNodeNotFound{}
}

// ErrStartNodeNotFound ошибка когда стартовая нода не найдена
type ErrStartNodeNotFound struct{}

func (e *ErrStartNodeNotFound) Error() string {
	return "start node not found in schema definition"
}

//
// isNilValue проверяет, является ли значение nil (включая typed nil, nil slice, nil map)
// TODO: вынести в утилиты
func isNilValue(data interface{}) bool {
    if data == nil {
        return true
    }
    
    value := reflect.ValueOf(data)
    switch value.Kind() {
    case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
        return value.IsNil()
    default:
        return false
    }
}

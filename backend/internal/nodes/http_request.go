package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPRequestConfig конфигурация http_request ноды
type HTTPRequestConfig struct {
	Method  string            `json:"method"`           // GET, POST, PUT, DELETE, etc
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	Timeout int               `json:"timeout"` // в секундах
	Retry   *HTTPRetryConfig  `json:"retry,omitempty"`
}

// HTTPRetryConfig настройки retry
type HTTPRetryConfig struct {
	MaxAttempts int   `json:"max_attempts"` // максимум попыток
	Delay       int   `json:"delay"`        // задержка между попытками (секунды)
	StatusCodes []int `json:"status_codes"` // коды для retry (например [500, 502, 503])
}

// HTTPRequestHandler обработчик HTTP запросов
type HTTPRequestHandler struct {
	client *http.Client
}

// NewHTTPRequestHandler создаёт новый HTTPRequestHandler
func NewHTTPRequestHandler() *HTTPRequestHandler {
	return &HTTPRequestHandler{
		client: &http.Client{
			Timeout: 30 * time.Second, // Дефолтный таймаут
		},
	}
}

// Execute выполняет http_request ноду
func (h *HTTPRequestHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config HTTPRequestConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse http_request config: %v", err)
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	// Устанавливаем timeout если указан
	if config.Timeout > 0 {
		h.client.Timeout = time.Duration(config.Timeout) * time.Second
	}

	// 1. Интерполируем URL
	url := InterpolateString(config.URL, execCtx)

	// 2. Подготавливаем body
	var bodyReader io.Reader
	if config.Body != nil {
		// Интерполируем body
		interpolatedBody := InterpolateValue(config.Body, execCtx)
		
		bodyJSON, err := json.Marshal(interpolatedBody)
		if err != nil {
			errMsg := fmt.Sprintf("failed to marshal body: %v", err)
			return &NodeResult{
				Status:     StatusFailed,
				Error:      &errMsg,
				ExitHandle: "error",
			}, nil
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	// 3. Создаём HTTP запрос
	req, err := http.NewRequestWithContext(ctx, config.Method, url, bodyReader)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	// 4. Добавляем headers
	if config.Headers != nil {
		interpolatedHeaders := InterpolateMap(config.Headers, execCtx)
		for k, v := range interpolatedHeaders {
			req.Header.Set(k, v)
		}
	}

	// Устанавливаем Content-Type для JSON body
	if config.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 5. Выполняем запрос с retry
	resp, err := h.doWithRetry(ctx, req, config.Retry)
	if err != nil {
		errMsg := fmt.Sprintf("request failed: %v", err)
		return &NodeResult{
			Output: map[string]interface{}{
				"error": err.Error(),
				"url":   url,
			},
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}
	defer resp.Body.Close()

	// 6. Читаем response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read response body: %v", err)
		return &NodeResult{
			Status:     StatusFailed,
			Error:      &errMsg,
			ExitHandle: "error",
		}, nil
	}

	// 7. Пытаемся распарсить как JSON
	var bodyJSON interface{}
	if err := json.Unmarshal(bodyBytes, &bodyJSON); err != nil {
		// Если не JSON - сохраняем как строку
		bodyJSON = string(bodyBytes)
	}

	// 8. Конвертируем headers в map
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 9. Определяем успешность по status code
	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	result := &NodeResult{
		Output: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     headers,
			"body":        bodyJSON,
			"raw_body":    string(bodyBytes),
		},
	}

	if isSuccess {
		result.Status = StatusSuccess
		result.ExitHandle = "success"
	} else {
		errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		result.Status = StatusFailed
		result.Error = &errMsg
		result.ExitHandle = "error"
	}

	return result, nil
}

// doWithRetry выполняет запрос с повторными попытками
func (h *HTTPRequestHandler) doWithRetry(ctx context.Context, req *http.Request, retry *HTTPRetryConfig) (*http.Response, error) {
	// Если retry не настроен - просто выполняем запрос
	if retry == nil || retry.MaxAttempts <= 1 {
		return h.client.Do(req)
	}

	var lastErr error
	var lastResp *http.Response

	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		// Клонируем запрос для повторной попытки
		reqClone := req.Clone(ctx)
		
		resp, err := h.client.Do(reqClone)
		
		// Если запрос успешен
		if err == nil {
			// Проверяем нужно ли retry для этого status code
			if !shouldRetry(resp.StatusCode, retry.StatusCodes) {
				return resp, nil
			}
			// Сохраняем response и закрываем body
			lastResp = resp
			resp.Body.Close()
		} else {
			lastErr = err
		}

		// Если это не последняя попытка - ждём
		if attempt < retry.MaxAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(retry.Delay) * time.Second):
				// Продолжаем
			}
		}
	}

	// Все попытки исчерпаны
	if lastResp != nil {
		return lastResp, nil // Возвращаем последний response
	}
	if lastErr != nil {
		return nil, fmt.Errorf("max retry attempts (%d) reached: %w", retry.MaxAttempts, lastErr)
	}
	return nil, fmt.Errorf("max retry attempts (%d) reached", retry.MaxAttempts)
}

// shouldRetry проверяет нужно ли повторить запрос для данного status code
func shouldRetry(statusCode int, retryCodes []int) bool {
	if len(retryCodes) == 0 {
		return false
	}
	
	for _, code := range retryCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}
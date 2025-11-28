package nodes

import (
	"context"
	"encoding/json"
	"fmt"
)

// HTTPRequestConfig конфигурация http_request ноды
type HTTPRequestConfig struct {
	Method  string                 `json:"method"`  // GET, POST, PUT, DELETE, etc
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers,omitempty"`
	Body    interface{}            `json:"body,omitempty"`
	Timeout int                    `json:"timeout"` // в секундах
	Retry   *HTTPRetryConfig       `json:"retry,omitempty"`
}

// HTTPRetryConfig настройки retry
type HTTPRetryConfig struct {
	MaxAttempts int   `json:"max_attempts"` // максимум попыток
	Delay       int   `json:"delay"`        // задержка между попытками (секунды)
	StatusCodes []int `json:"status_codes"` // коды для retry (например [500, 502, 503])
}

// HTTPRequestHandler обработчик HTTP запросов
type HTTPRequestHandler struct {
	// TODO: добавить http.Client
}

// NewHTTPRequestHandler создаёт новый HTTPRequestHandler
func NewHTTPRequestHandler() *HTTPRequestHandler {
	return &HTTPRequestHandler{
		// TODO: инициализировать http.Client с timeout
	}
}

// Execute выполняет http_request ноду
func (h *HTTPRequestHandler) Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error) {
	var config HTTPRequestConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errMsg := fmt.Sprintf("failed to parse http_request config: %v", err)
		return &NodeResult{
			Status: StatusFailed,
			Error:  &errMsg,
		}, nil
	}

	// TODO: Реализовать:
	// 1. Интерполировать переменные в URL, headers, body
	//    url := interpolateVariables(config.URL, execCtx)
	//
	// 2. Построить HTTP запрос
	//    req, err := http.NewRequestWithContext(ctx, config.Method, url, body)
	//
	// 3. Добавить headers
	//    for k, v := range config.Headers {
	//        req.Header.Set(k, interpolateVariables(v, execCtx))
	//    }
	//
	// 4. Выполнить с retry
	//    resp, err := h.doWithRetry(ctx, req, config.Retry)
	//
	// 5. Прочитать response body
	//    bodyBytes, err := io.ReadAll(resp.Body)
	//    defer resp.Body.Close()
	//
	// 6. Попробовать распарсить как JSON
	//    var bodyJSON interface{}
	//    json.Unmarshal(bodyBytes, &bodyJSON)
	//
	// 7. Вернуть результат
	//    return &NodeResult{
	//        Output: map[string]interface{}{
	//            "status_code": resp.StatusCode,
	//            "headers": resp.Header,
	//            "body": bodyJSON,
	//            "raw_body": string(bodyBytes),
	//        },
	//        Status: StatusSuccess,
	//    }, nil

	// Заглушка
	errMsg := "http_request node not implemented yet"
	return &NodeResult{
		Status: StatusFailed,
		Error:  &errMsg,
	}, nil
}

// TODO: Реализовать doWithRetry
// func (h *HTTPRequestHandler) doWithRetry(ctx context.Context, req *http.Request, retry *HTTPRetryConfig) (*http.Response, error) {
//     if retry == nil {
//         return h.client.Do(req)
//     }
//
//     var lastErr error
//     for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
//         resp, err := h.client.Do(req)
//         if err == nil {
//             // Проверяем нужно ли retry для этого status code
//             if !shouldRetry(resp.StatusCode, retry.StatusCodes) {
//                 return resp, nil
//             }
//             resp.Body.Close()
//         }
//         lastErr = err
//         
//         if attempt < retry.MaxAttempts {
//             time.Sleep(time.Duration(retry.Delay) * time.Second)
//         }
//     }
//     return nil, fmt.Errorf("max retry attempts reached: %w", lastErr)
// }

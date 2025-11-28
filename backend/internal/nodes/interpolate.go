package nodes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Регулярка для поиска {{variable_name}}
	varPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)
)

// InterpolateString заменяет {{variable}} на значения из контекста
func InterpolateString(s string, ctx *ExecutionContext) string {
	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Извлекаем имя переменной из {{...}}
		varName := strings.TrimSpace(match[2 : len(match)-2])
		
		// Ищем значение в Variables
		if val, ok := ctx.Variables[varName]; ok {
			return fmt.Sprintf("%v", val)
		}
		
		// Если не нашли - возвращаем как есть
		return match
	})
}

// InterpolateValue интерполирует значение (может быть строка, map, slice)
func InterpolateValue(value interface{}, ctx *ExecutionContext) interface{} {
	switch v := value.(type) {
	case string:
		return InterpolateString(v, ctx)
		
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = InterpolateValue(val, ctx)
		}
		return result
		
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = InterpolateValue(val, ctx)
		}
		return result
		
	default:
		return value
	}
}

// InterpolateMap интерполирует все строковые значения в map
func InterpolateMap(m map[string]string, ctx *ExecutionContext) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = InterpolateString(v, ctx)
	}
	return result
}

// ResolveVariableOrValue получает значение - либо из Variables, либо возвращает как есть
// Используется для операндов в math ноде
func ResolveVariableOrValue(value interface{}, ctx *ExecutionContext) interface{} {
	// Если это строка - проверяем, не имя ли это переменной
	if str, ok := value.(string); ok {
		// Сначала проверяем есть ли {{...}}
		if strings.Contains(str, "{{") {
			return InterpolateString(str, ctx)
		}
		
		// Если нет {{...}}, но есть такая переменная - возвращаем её значение
		if val, exists := ctx.Variables[str]; exists {
			return val
		}
	}
	return value
}

// MarshalInterpolated маршалит значение в JSON и интерполирует
func MarshalInterpolated(value interface{}, ctx *ExecutionContext) ([]byte, error) {
	// Сначала маршалим в JSON
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	
	// Потом интерполируем строку
	interpolated := InterpolateString(string(jsonBytes), ctx)
	
	return []byte(interpolated), nil
}
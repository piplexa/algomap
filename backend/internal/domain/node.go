package domain

import "encoding/json"

// NodeType константы типов нод
const (
	NodeTypeStart          = "start"
	NodeTypeEnd            = "end"
	NodeTypeCondition      = "condition"
	NodeTypeHTTPRequest    = "http_request"
	NodeTypeLog            = "log"
	NodeTypeSleep          = "sleep"
	NodeTypeVariableSet    = "variable_set"
	NodeTypeMath           = "math"
	NodeTypeRabbitMQPublish = "rabbitmq_publish"
	NodeTypeSubSchema      = "sub_schema" // TODO: будет реализовано позже
)

// NodeConfig базовая структура для конфигурации ноды
type NodeConfig struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

// TODO: Добавить специфичные конфиги для каждого типа нод:
// - StartConfig
// - EndConfig
// - ConditionConfig
// - HTTPRequestConfig
// - LogConfig
// - SleepConfig
// и т.д.

// TODO: Реализовать в worker'е
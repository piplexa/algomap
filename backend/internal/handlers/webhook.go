package handlers

// WebhookHandler - HTTP handlers для webhook триггеров
// Судя по всему лишние, так как есть executions/...../continue

// TODO: Реализовать endpoints:
// POST   /webhook/:schema_id              - запуск через webhook
// POST   /webhook/:schema_id/:step_id     - запуск с конкретного шага
//
// Payload из body должен быть доступен внутри схемы как {{webhook.payload.*}}

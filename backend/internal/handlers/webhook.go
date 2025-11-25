package handlers

// WebhookHandler - HTTP handlers для webhook триггеров

// TODO: Реализовать endpoints:
// POST   /webhook/:schema_id              - запуск через webhook
// POST   /webhook/:schema_id/:step_id     - запуск с конкретного шага
//
// Payload из body должен быть доступен внутри схемы как {{webhook.payload.*}}
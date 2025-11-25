package executor

// ExecutionContext - контекст выполнения схемы
// Содержит все переменные, доступные внутри схемы

// TODO: Реализовать в worker'е:
// - webhook.payload.*
// - user.email
// - execution.id
// - steps.<node_id>.output.*
// - variables.*
// - env.*
//
// Интерполяция переменных: {{path.to.variable}}
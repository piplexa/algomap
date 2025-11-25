package nodes

// NodeHandler - интерфейс для обработчиков нод
// Каждый тип ноды реализует этот интерфейс

// TODO: Реализовать в worker'е:
//
// type NodeHandler interface {
//     Execute(ctx context.Context, node *Node, execCtx *ExecutionContext) (*NodeResult, error)
// }
//
// type NodeResult struct {
//     Output    map[string]interface{}  // Результат выполнения
//     NextNodes []string                // Куда переходить дальше
//     Sleep     *time.Duration          // Если нужна задержка
//     Error     error                   // Ошибка выполнения
// }
//
// Реализовать для типов нод:
// - StartHandler
// - EndHandler
// - ConditionHandler
// - HTTPRequestHandler
// - LogHandler
// - SleepHandler
// - VariableSetHandler
// - MathHandler
// - RabbitMQPublishHandler
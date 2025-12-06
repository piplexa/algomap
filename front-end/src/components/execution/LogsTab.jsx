export default function LogsTab({ steps, selectedStep }) {
  const formatTime = (isoString) => {
    if (!isoString) return '';
    const date = new Date(isoString);
    return date.toLocaleTimeString('ru-RU', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    });
  };

  // Если выбран конкретный шаг, показываем только его логи
  // Иначе показываем все логи
  const logsSteps = selectedStep
    ? steps.filter((step) => step.id === selectedStep && step.node_type === 'log' && step.log_message)
    : steps.filter((step) => step.node_type === 'log' && step.log_message);

  return (
    <div className="logs-list">
      {logsSteps.length === 0 && (
        <div className="empty-state">
          {selectedStep ? 'У этого шага нет логов' : 'Нет логов'}
        </div>
      )}
      {logsSteps.map((step) => (
        <div key={step.id} className="log-entry">
          <span className="log-time">{formatTime(step.started_at)}</span>
          <span className="log-message">{step.log_message}</span>
        </div>
      ))}
    </div>
  );
}

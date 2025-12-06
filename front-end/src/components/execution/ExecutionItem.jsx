export default function ExecutionItem({ execution, isSelected, onClick }) {
  const formatDateTime = (isoString) => {
    const date = new Date(isoString);
    return date.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const formatDuration = (ms) => {
    if (ms < 1000) return `${ms}мс`;
    return `${(ms / 1000).toFixed(1)}с`;
  };

  return (
    <div
      className={`execution-item ${isSelected ? 'selected' : ''}`}
      onClick={onClick}
    >
      <div className="execution-time">{formatDateTime(execution.created_at)}</div>
      <div className={`execution-status ${execution.status}`}>
        {execution.status === 'completed' ? '✓ Успех' : '✗ Ошибка'}
      </div>
      <div className="execution-duration">
        {execution.cnt_executed_steps } шагов | {formatDuration(execution.duration)}
      </div>
    </div>
  );
}

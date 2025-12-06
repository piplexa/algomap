import { NODE_DEFINITIONS } from '../../utils/nodeTypes';

export default function StepsTab({ steps, selectedStep, onSelectStep }) {
  const formatDuration = (step) => {
    if (!step.started_at || !step.finished_at) return '';
    const start = new Date(step.started_at);
    const end = new Date(step.finished_at);
    const duration = end - start;
    if (duration < 1000) return `${duration}мс`;
    return `${(duration / 1000).toFixed(1)}с`;
  };

  const getNodeIcon = (nodeType) => {
    const definition = NODE_DEFINITIONS[nodeType];
    return definition?.icon || '○';
  };

  return (
    <div className="steps-list">
      {steps.length === 0 && (
        <div className="empty-state">Нет шагов</div>
      )}
      {steps.map((step) => (
        <div
          key={step.id}
          className={`step-row ${selectedStep === step.id ? 'selected' : ''}`}
          onClick={() => onSelectStep(step.id)}
        >
          <div className="step-row-header">
            <span className="step-icon">{getNodeIcon(step.node_type)}</span>
            <span className={`step-status ${step.status}`}>
              {step.status === 'success' ? '✓' : '✗'}
            </span>
          </div>
          <div className="step-name">{step.node_id}</div>
          <div className="step-time">{formatDuration(step)}</div>
        </div>
      ))}
    </div>
  );
}

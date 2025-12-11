import { useState, useEffect } from 'react';
import '../styles/EdgeConfigPanel.css';

// Доступные типы линий в React Flow
const EDGE_TYPES = [
  { value: 'smoothstep', label: 'Smooth Step', description: 'Сглаженные ступеньки (рекомендуется)' },
  { value: 'step', label: 'Step', description: 'Прямые углы' },
  { value: 'straight', label: 'Straight', description: 'Прямая линия' },
  { value: 'default', label: 'Bezier', description: 'Кривая Безье' },
];

export default function EdgeConfigPanel({ edge, onUpdate, onClose }) {
  const [edgeType, setEdgeType] = useState(edge?.type || 'smoothstep');
  const [animated, setAnimated] = useState(edge?.animated || false);

  // Обновляем локальное состояние при изменении выбранной линии
  useEffect(() => {
    if (edge) {
      setEdgeType(edge.type || 'smoothstep');
      setAnimated(edge.animated || false);
    }
  }, [edge]);

  if (!edge) {
    return (
      <div className="edge-config-panel empty">
        <div className="empty-state">
          <span className="empty-icon">🔗</span>
          <p>Выберите линию соединения</p>
          <small>Кликните на линию между блоками</small>
        </div>
      </div>
    );
  }

  // Обновление типа линии
  const handleTypeChange = (newType) => {
    setEdgeType(newType);
    onUpdate(edge.id, { type: newType });
  };

  // Обновление анимации
  const handleAnimatedChange = (newAnimated) => {
    setAnimated(newAnimated);
    onUpdate(edge.id, { animated: newAnimated });
  };

  return (
    <div className="edge-config-panel">
      <div className="panel-header">
        <div>
          <span className="panel-icon">🔗</span>
          <h3>Настройки линии</h3>
        </div>
        <button onClick={onClose} className="close-btn">
          ×
        </button>
      </div>

      <div className="panel-body">
        <div className="form-section">
          <label className="form-label">ID соединения</label>
          <div className="edge-id">{edge.id}</div>
        </div>

        <div className="form-section">
          <label className="form-label">Тип линии</label>
          <div className="edge-types">
            {EDGE_TYPES.map((type) => (
              <div
                key={type.value}
                className={`edge-type-option ${edgeType === type.value ? 'selected' : ''}`}
                onClick={() => handleTypeChange(type.value)}
              >
                <div className="edge-type-header">
                  <input
                    type="radio"
                    checked={edgeType === type.value}
                    onChange={() => handleTypeChange(type.value)}
                    className="edge-type-radio"
                  />
                  <span className="edge-type-label">{type.label}</span>
                </div>
                <p className="edge-type-description">{type.description}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="form-section">
          <label className="form-label">
            <input
              type="checkbox"
              checked={animated}
              onChange={(e) => handleAnimatedChange(e.target.checked)}
              className="checkbox"
            />
            <span>Анимация движения</span>
          </label>
          <p className="form-help">
            Анимированные пунктирные линии показывают направление потока
          </p>
        </div>

        <div className="form-section">
          <label className="form-label">Информация</label>
          <div className="edge-info">
            <div className="info-row">
              <span className="info-label">От блока:</span>
              <span className="info-value">{edge.source}</span>
            </div>
            <div className="info-row">
              <span className="info-label">К блоку:</span>
              <span className="info-value">{edge.target}</span>
            </div>
          </div>
        </div>

        <div className="panel-tips">
          <h4>💡 Рекомендации:</h4>
          <ul>
            <li><strong>Smooth Step</strong> - лучше всего обходит блоки</li>
            <li><strong>Step</strong> - для строгих прямоугольных схем</li>
            <li><strong>Straight</strong> - для минимализма</li>
            <li><strong>Bezier</strong> - классический вид блок-схем</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import HistoryMode from './execution/HistoryMode';
import DebugMode from './execution/DebugMode';
import '../styles/ExecutionPanel.css';

export default function ExecutionPanel({ schemaId, isOpen, onToggle, onNodeHighlight }) {
  const [mode, setMode] = useState('history'); // 'history' | 'debug'
  const [height, setHeight] = useState(400);

  // Загрузка сохраненной высоты панели из localStorage
  useEffect(() => {
    const savedHeight = localStorage.getItem('executionPanelHeight');
    if (savedHeight) {
      setHeight(parseInt(savedHeight, 10));
    }
  }, []);

  // Сохранение высоты панели в localStorage
  const handleHeightChange = (newHeight) => {
    setHeight(newHeight);
    localStorage.setItem('executionPanelHeight', newHeight.toString());
  };

  return (
    <div
      className={`execution-panel ${isOpen ? 'open' : 'closed'}`}
      style={{ height: isOpen ? `${height}px` : '40px' }}
    >
      {/* Заголовок с переключателем режимов */}
      <div className="execution-panel-header">
        <button
          className={`mode-btn ${mode === 'history' ? 'active' : ''}`}
          onClick={() => {
            setMode('history');
            if (!isOpen) {
              onToggle();
            }
          }}
        >
          История
        </button>
        <button
          className={`mode-btn ${mode === 'debug' ? 'active' : ''}`}
          onClick={() => {
            setMode('debug');
            if (!isOpen) {
              onToggle();
            }
          }}
        >
          Отладка
        </button>
        <button className="panel-toggle" onClick={onToggle}>
          {isOpen ? '▼' : '▲'}
        </button>
      </div>

      {/* Контент в зависимости от режима */}
      {isOpen && (
        <div className="execution-panel-content">
          {mode === 'history' ? (
            <HistoryMode schemaId={schemaId} onNodeHighlight={onNodeHighlight} />
          ) : (
            <DebugMode schemaId={schemaId} onNodeHighlight={onNodeHighlight} />
          )}
        </div>
      )}
    </div>
  );
}

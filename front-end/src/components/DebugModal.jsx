import { useState } from 'react';
import '../styles/DebugModal.css';

export default function DebugModal({ schema, onClose }) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [variables, setVariables] = useState({
    webhook: { payload: { example: 'test_value' } },
    user: { email: 'user@example.com' },
    execution: { id: 'debug-exec-001' },
    steps: {},
    variables: {},
  });
  const [logs, setLogs] = useState([]);

  // Получаем путь выполнения (упрощённо - по порядку нод)
  const executionPath = schema.definition?.nodes || [];
  const currentNode = executionPath[currentStep];

  const addLog = (type, message) => {
    setLogs(prev => [...prev, { 
      type, 
      message, 
      timestamp: new Date().toLocaleTimeString(),
      step: currentStep 
    }]);
  };

  const executeStep = () => {
    if (currentStep >= executionPath.length) {
      addLog('success', '✅ Выполнение завершено');
      setIsRunning(false);
      return;
    }

    const node = executionPath[currentStep];
    addLog('info', `▶️ Выполняется: ${node.data.label} (${node.id})`);

    // Симуляция выполнения в зависимости от типа ноды
    setTimeout(() => {
      switch (node.data.type) {
        case 'start':
          addLog('info', '🚀 Схема запущена');
          break;

        case 'log':
          const logMsg = node.data.config.message || 'Empty log';
          addLog('info', `📝 Лог: ${logMsg}`);
          break;

        case 'http_request':
          const url = node.data.config.url || 'http://example.com';
          addLog('info', `🌐 HTTP ${node.data.config.method} ${url}`);
          // Симуляция ответа
          const mockResponse = { status: 200, data: { result: 'ok' } };
          setVariables(prev => ({
            ...prev,
            steps: {
              ...prev.steps,
              [node.id]: { output: mockResponse }
            }
          }));
          addLog('success', `✅ Получен ответ: ${JSON.stringify(mockResponse)}`);
          break;

        case 'variable_set':
          const varName = node.data.config.variable;
          const varValue = node.data.config.value;
          setVariables(prev => ({
            ...prev,
            variables: {
              ...prev.variables,
              [varName]: varValue
            }
          }));
          addLog('success', `💾 Переменная установлена: ${varName} = ${varValue}`);
          break;

        case 'condition':
          const expr = node.data.config.expression || 'true';
          addLog('info', `🔀 Условие: ${expr}`);
          addLog('success', `✅ Результат: true (для отладки)`);
          break;

        case 'sleep':
          const duration = node.data.config.duration;
          addLog('info', `⏰ Задержка ${duration} ${node.data.config.unit}`);
          break;

        case 'end':
          addLog('success', `⏹️ ${node.data.config.message || 'Завершено'}`);
          break;

        default:
          addLog('info', `▶️ Выполнена нода типа: ${node.data.type}`);
      }

      setCurrentStep(prev => prev + 1);
    }, 500); // Симуляция задержки
  };

  const handleNext = () => {
    executeStep();
  };

  const handlePlay = () => {
    setIsRunning(true);
    const interval = setInterval(() => {
      if (currentStep >= executionPath.length - 1) {
        clearInterval(interval);
        setIsRunning(false);
      } else {
        executeStep();
      }
    }, 1000);
  };

  const handleStop = () => {
    setIsRunning(false);
    setCurrentStep(0);
    setLogs([]);
    setVariables({
      webhook: { payload: { example: 'test_value' } },
      user: { email: 'user@example.com' },
      execution: { id: 'debug-exec-001' },
      steps: {},
      variables: {},
    });
  };

  return (
    <div className="debug-modal-overlay" onClick={onClose}>
      <div className="debug-modal" onClick={(e) => e.stopPropagation()}>
        <div className="debug-header">
          <h2>🐛 Отладка: {schema.name}</h2>
          <button onClick={onClose} className="close-btn">✕</button>
        </div>

        <div className="debug-content">
          {/* Левая панель - Схема с подсветкой */}
          <div className="debug-schema">
            <h3>Схема</h3>
            <div className="debug-nodes">
              {executionPath.map((node, index) => (
                <div
                  key={node.id}
                  className={`debug-node ${index === currentStep ? 'active' : ''} ${index < currentStep ? 'completed' : ''}`}
                >
                  <span className="node-index">{index + 1}</span>
                  <span className="node-label">{node.data.label}</span>
                  <span className="node-id">{node.id}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Центральная панель - Логи */}
          <div className="debug-logs">
            <h3>Логи выполнения</h3>
            <div className="logs-container">
              {logs.map((log, i) => (
                <div key={i} className={`log-entry log-${log.type}`}>
                  <span className="log-time">{log.timestamp}</span>
                  <span className="log-message">{log.message}</span>
                </div>
              ))}
              {logs.length === 0 && (
                <div className="log-empty">Нажмите "Шаг вперёд" для начала</div>
              )}
            </div>
          </div>

          {/* Правая панель - Переменные */}
          <div className="debug-variables">
            <h3>Переменные</h3>
            <div className="variables-container">
              <div className="variable-group">
                <strong>webhook.payload:</strong>
                <pre>{JSON.stringify(variables.webhook.payload, null, 2)}</pre>
              </div>
              
              <div className="variable-group">
                <strong>user:</strong>
                <pre>{JSON.stringify(variables.user, null, 2)}</pre>
              </div>

              {Object.keys(variables.variables).length > 0 && (
                <div className="variable-group">
                  <strong>variables:</strong>
                  <pre>{JSON.stringify(variables.variables, null, 2)}</pre>
                </div>
              )}

              {Object.keys(variables.steps).length > 0 && (
                <div className="variable-group">
                  <strong>steps:</strong>
                  <pre>{JSON.stringify(variables.steps, null, 2)}</pre>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Контролы */}
        <div className="debug-controls">
          <div className="debug-progress">
            Шаг: {currentStep + 1} / {executionPath.length}
          </div>
          <div className="debug-buttons">
            <button 
              onClick={handleNext} 
              disabled={isRunning || currentStep >= executionPath.length}
              className="btn-primary"
            >
              ▶️ Шаг вперёд
            </button>
            <button 
              onClick={handlePlay} 
              disabled={isRunning || currentStep >= executionPath.length}
              className="btn-success"
            >
              ⏩ Автозапуск
            </button>
            <button 
              onClick={handleStop}
              className="btn-danger"
            >
              ⏹️ Сброс
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

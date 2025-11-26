import { useState, useEffect } from 'react';
import { NODE_DEFINITIONS } from '../utils/nodeTypes';

export default function NodeConfigPanel({ node, onUpdate, onClose }) {
  const [config, setConfig] = useState(node?.data?.config || {});
  const definition = NODE_DEFINITIONS[node?.data?.type];

  useEffect(() => {
    if (node) {
      setConfig(node.data.config || {});
    }
  }, [node]);

  if (!node) {
    return (
      <div className="node-config-panel empty">
        <p>Выберите блок для настройки</p>
      </div>
    );
  }

  const handleChange = (key, value) => {
    const newConfig = { ...config, [key]: value };
    setConfig(newConfig);
    onUpdate(node.id, newConfig);
  };

  const handleNestedChange = (parentKey, childKey, value) => {
    const newConfig = {
      ...config,
      [parentKey]: {
        ...config[parentKey],
        [childKey]: value,
      },
    };
    setConfig(newConfig);
    onUpdate(node.id, newConfig);
  };

  return (
    <div className="node-config-panel">
      <div className="panel-header">
        <div>
          <span className="panel-icon">{definition?.icon}</span>
          <h3>{definition?.label}</h3>
        </div>
        <button onClick={onClose} className="close-btn">✕</button>
      </div>

      <div className="panel-body">
        <div className="form-group">
          <label>ID ноды</label>
          <input type="text" value={node.id} disabled />
        </div>

        {renderConfigFields(node.data.type, config, handleChange, handleNestedChange)}
      </div>
    </div>
  );
}

function renderConfigFields(nodeType, config, handleChange, handleNestedChange) {
  switch (nodeType) {
    case 'start':
      return <p className="info-text">Стартовая нода не требует настройки</p>;

    case 'end':
      return (
        <>
          <div className="form-group">
            <label>
              <input
                type="checkbox"
                checked={config.success}
                onChange={(e) => handleChange('success', e.target.checked)}
              />
              Успешное завершение
            </label>
          </div>
          <div className="form-group">
            <label>Сообщение</label>
            <input
              type="text"
              value={config.message || ''}
              onChange={(e) => handleChange('message', e.target.value)}
              placeholder="Execution completed"
            />
          </div>
        </>
      );

    case 'log':
      return (
        <>
          <div className="form-group">
            <label>Уровень логирования</label>
            <select
              value={config.level || 'info'}
              onChange={(e) => handleChange('level', e.target.value)}
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warning</option>
              <option value="error">Error</option>
            </select>
          </div>
          <div className="form-group">
            <label>Сообщение</label>
            <textarea
              value={config.message || ''}
              onChange={(e) => handleChange('message', e.target.value)}
              placeholder="Processing user {{user.email}}"
              rows={3}
            />
            <small>Используйте {'{{'} {'}}'}  для переменных</small>
          </div>
        </>
      );

    case 'http_request':
      return (
        <>
          <div className="form-group">
            <label>Метод</label>
            <select
              value={config.method || 'GET'}
              onChange={(e) => handleChange('method', e.target.value)}
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
              <option value="PATCH">PATCH</option>
            </select>
          </div>
          <div className="form-group">
            <label>URL</label>
            <input
              type="text"
              value={config.url || ''}
              onChange={(e) => handleChange('url', e.target.value)}
              placeholder="https://api.example.com/users"
            />
          </div>
          <div className="form-group">
            <label>Timeout (сек)</label>
            <input
              type="number"
              value={config.timeout || 30}
              onChange={(e) => handleChange('timeout', parseInt(e.target.value))}
              min={1}
            />
          </div>
          <div className="form-group">
            <label>
              <input
                type="checkbox"
                checked={config.retry?.enabled || false}
                onChange={(e) =>
                  handleNestedChange('retry', 'enabled', e.target.checked)
                }
              />
              Включить retry
            </label>
          </div>
        </>
      );

    case 'condition':
      return (
        <div className="form-group">
          <label>Условие</label>
          <textarea
            value={config.expression || ''}
            onChange={(e) => handleChange('expression', e.target.value)}
            placeholder="{{webhook.payload.age}} > 18"
            rows={3}
          />
          <small>Например: {'{{'} webhook.payload.age {'}}'}  {'>'} 18</small>
        </div>
      );

    case 'variable_set':
      return (
        <>
          <div className="form-group">
            <label>Имя переменной</label>
            <input
              type="text"
              value={config.variable || ''}
              onChange={(e) => handleChange('variable', e.target.value)}
              placeholder="user_age"
            />
          </div>
          <div className="form-group">
            <label>Значение</label>
            <input
              type="text"
              value={config.value || ''}
              onChange={(e) => handleChange('value', e.target.value)}
              placeholder="{{webhook.payload.age}}"
            />
          </div>
        </>
      );

    case 'sleep':
      return (
        <>
          <div className="form-group">
            <label>Длительность</label>
            <input
              type="number"
              value={config.duration || 60}
              onChange={(e) => handleChange('duration', parseInt(e.target.value))}
              min={1}
            />
          </div>
          <div className="form-group">
            <label>Единица измерения</label>
            <select
              value={config.unit || 'seconds'}
              onChange={(e) => handleChange('unit', e.target.value)}
            >
              <option value="seconds">Секунды</option>
              <option value="minutes">Минуты</option>
              <option value="hours">Часы</option>
              <option value="days">Дни</option>
            </select>
          </div>
        </>
      );

    case 'math':
      return (
        <>
          <div className="form-group">
            <label>Операция</label>
            <select
              value={config.operation || 'add'}
              onChange={(e) => handleChange('operation', e.target.value)}
            >
              <option value="add">Сложение (+)</option>
              <option value="subtract">Вычитание (-)</option>
              <option value="multiply">Умножение (*)</option>
              <option value="divide">Деление (/)</option>
              <option value="modulo">Остаток (%)</option>
            </select>
          </div>
          <div className="form-group">
            <label>Операнд 1</label>
            <input
              type="text"
              value={config.operand1 || ''}
              onChange={(e) => handleChange('operand1', e.target.value)}
              placeholder="10 или {{variables.price}}"
            />
          </div>
          <div className="form-group">
            <label>Операнд 2</label>
            <input
              type="text"
              value={config.operand2 || ''}
              onChange={(e) => handleChange('operand2', e.target.value)}
              placeholder="5"
            />
          </div>
          <div className="form-group">
            <label>Сохранить результат в</label>
            <input
              type="text"
              value={config.result_variable || ''}
              onChange={(e) => handleChange('result_variable', e.target.value)}
              placeholder="final_price"
            />
          </div>
        </>
      );

    case 'rabbitmq_publish':
      return (
        <>
          <div className="form-group">
            <label>Очередь</label>
            <input
              type="text"
              value={config.queue || ''}
              onChange={(e) => handleChange('queue', e.target.value)}
              placeholder="notifications"
            />
          </div>
          <div className="form-group">
            <label>Exchange (опционально)</label>
            <input
              type="text"
              value={config.exchange || ''}
              onChange={(e) => handleChange('exchange', e.target.value)}
            />
          </div>
        </>
      );

    default:
      return <p className="info-text">Нет настроек для этого типа ноды</p>;
  }
}

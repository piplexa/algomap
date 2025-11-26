import { Handle, Position } from 'reactflow';
import { NODE_DEFINITIONS } from '../utils/nodeTypes';

export default function CustomNode({ data, isConnectable }) {
  const definition = NODE_DEFINITIONS[data.type];

  return (
    <div 
      className={`custom-node ${data.selected ? 'selected' : ''}`}
      style={{ borderColor: definition?.color }}
    >
      <Handle
        type="target"
        position={Position.Top}
        isConnectable={isConnectable}
      />

      <div className="node-header" style={{ backgroundColor: definition?.color }}>
        <span className="node-icon">{definition?.icon}</span>
        <span className="node-label">{definition?.label}</span>
      </div>

      <div className="node-body">
        <div className="node-id">{data.id}</div>
        {data.config && Object.keys(data.config).length > 0 && (
          <div className="node-preview">
            {getConfigPreview(data.type, data.config)}
          </div>
        )}
      </div>

      {definition?.outputs > 0 && (
        <>
          {definition.outputs === 1 ? (
            <Handle
              type="source"
              position={Position.Bottom}
              id="output"
              isConnectable={isConnectable}
            />
          ) : (
            <>
              <Handle
                type="source"
                position={Position.Bottom}
                id="success"
                style={{ left: '33%' }}
                isConnectable={isConnectable}
              />
              <Handle
                type="source"
                position={Position.Bottom}
                id="error"
                style={{ left: '66%' }}
                isConnectable={isConnectable}
              />
            </>
          )}
        </>
      )}
    </div>
  );
}

function getConfigPreview(type, config) {
  switch (type) {
    case 'http_request':
      return `${config.method} ${config.url || '...'}`;
    case 'condition':
      return config.expression || 'Условие не задано';
    case 'log':
      return config.message || 'Лог пустой';
    case 'variable_set':
      return `${config.variable} = ${config.value}`;
    case 'sleep':
      return `${config.duration} ${config.unit}`;
    case 'math':
      return `${config.operation}(${config.operand1}, ${config.operand2})`;
    default:
      return null;
  }
}

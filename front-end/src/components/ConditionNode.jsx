import { Handle, Position } from 'reactflow';
import { NODE_DEFINITIONS } from '../utils/nodeTypes';

export default function ConditionNode({ data, isConnectable }) {
  const definition = NODE_DEFINITIONS[data.type];

  return (
    <div className="condition-node-wrapper">
      <Handle
        type="target"
        position={Position.Top}
        isConnectable={isConnectable}
      />

      <div
        className={`condition-node ${data.selected ? 'selected' : ''}`}
        style={{ borderColor: definition?.color }}
      >
        <div className="condition-header" style={{ backgroundColor: definition?.color, display: 'none' }}>
          <span className="node-icon">{definition?.icon}</span>
          <span className="node-label">{definition?.label}</span>
        </div>

        <div className="condition-body">
          {data.config?.expression ? (
            <div className="condition-expression">{data.config.expression}</div>
          ) : (
            <div className="condition-placeholder">Условие не задано</div>
          )}
        </div>
      </div>

      {/* Два выхода: true (справа, зелёный) и false (слева, красный) */}
      <Handle
        type="source"
        position={Position.Right}
        id="true"
        style={{
          background: '#10b981',
          width: '18px',
          height: '18px',
          border: '2px solid white'
        }}
        isConnectable={isConnectable}
      />
      <Handle
        type="source"
        position={Position.Left}
        id="false"
        style={{
          background: '#ef4444',
          width: '18px',
          height: '18px',
          border: '2px solid white'
        }}
        isConnectable={isConnectable}
      />
    </div>
  );
}
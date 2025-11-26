import { getNodesByCategory } from '../utils/nodeTypes';

export default function NodesPalette() {
  const nodesByCategory = getNodesByCategory();

  const onDragStart = (event, nodeType) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.effectAllowed = 'move';
  };

  return (
    <div className="nodes-palette">
      <h3>Блоки</h3>
      {Object.entries(nodesByCategory).map(([category, nodes]) => (
        <div key={category} className="palette-category">
          <h4>{category}</h4>
          <div className="palette-nodes">
            {nodes.map((node) => (
              <div
                key={node.type}
                className="palette-node"
                draggable
                onDragStart={(e) => onDragStart(e, node.type)}
                style={{ borderLeftColor: node.color }}
              >
                <span className="palette-node-icon">{node.icon}</span>
                <span className="palette-node-label">{node.label}</span>
              </div>
            ))}
          </div>
        </div>
      ))}

      <div className="palette-help">
        <p>💡 <strong>Как использовать:</strong></p>
        <p>Перетащите блок на холст и соедините с другими блоками</p>
      </div>
    </div>
  );
}

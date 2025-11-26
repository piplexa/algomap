import { useState, useCallback, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
} from 'reactflow';
import 'reactflow/dist/style.css';

import CustomNode from '../components/CustomNode';
import NodesPalette from '../components/NodesPalette';
import NodeConfigPanel from '../components/NodeConfigPanel';
import DebugModal from '../components/DebugModal';
import { useSchemasStore } from '../store/schemasStore';
import { NODE_DEFINITIONS } from '../utils/nodeTypes';
import '../styles/Editor.css';

const nodeTypes = {
  custom: CustomNode,
};

let nodeIdCounter = 1;

export default function EditorPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const reactFlowWrapper = useRef(null);
  const [reactFlowInstance, setReactFlowInstance] = useState(null);

  const { currentSchema, fetchSchemaById, updateSchema } = useSchemasStore();
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState(null);
  const [schemaName, setSchemaName] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [showDebugModal, setShowDebugModal] = useState(false);

  // Загрузка схемы
  useEffect(() => {
    if (id) {
      fetchSchemaById(id);
    }
  }, [id, fetchSchemaById]);

  // Заполнение редактора данными
  useEffect(() => {
    if (currentSchema) {
      setSchemaName(currentSchema.name);
      
      if (currentSchema.definition?.nodes) {
        const loadedNodes = currentSchema.definition.nodes.map((node) => ({
          ...node,
          type: 'custom',
          data: {
            ...node.data,
            selected: false,
          },
        }));
        setNodes(loadedNodes);
        
        // Обновляем счетчик ID
        const maxId = Math.max(...loadedNodes.map(n => {
          const match = n.id.match(/_(\d+)$/);
          return match ? parseInt(match[1]) : 0;
        }), 0);
        nodeIdCounter = maxId + 1;
      }

      if (currentSchema.definition?.edges) {
        setEdges(currentSchema.definition.edges);
      }
    }
  }, [currentSchema, setNodes, setEdges]);

  // Drag & Drop ноды с палитры
  const onDragOver = useCallback((event) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event) => {
      event.preventDefault();

      const nodeType = event.dataTransfer.getData('application/reactflow');
      if (!nodeType || !reactFlowInstance) return;

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const definition = NODE_DEFINITIONS[nodeType];
      const newNode = {
        id: `${nodeType}_${nodeIdCounter++}`,
        type: 'custom',
        position,
        data: {
          type: nodeType,
          label: definition.label,
          config: { ...definition.config },
          selected: false,
        },
      };

      setNodes((nds) => nds.concat(newNode));
    },
    [reactFlowInstance, setNodes]
  );

  // Соединение нод
  const onConnect = useCallback(
    (params) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  // Выбор ноды
  const onNodeClick = useCallback(
    (event, node) => {
      setNodes((nds) =>
        nds.map((n) => ({
          ...n,
          data: { ...n.data, selected: n.id === node.id },
        }))
      );
      setSelectedNode(node);
    },
    [setNodes]
  );

  // Обновление конфига ноды
  const onNodeConfigUpdate = useCallback(
    (nodeId, newConfig) => {
      setNodes((nds) =>
        nds.map((node) => {
          if (node.id === nodeId) {
            return {
              ...node,
              data: {
                ...node.data,
                config: newConfig,
              },
            };
          }
          return node;
        })
      );
    },
    [setNodes]
  );

  // Удаление ноды
  const onNodesDelete = useCallback((deleted) => {
    // Если удалили выбранную ноду, сбрасываем выбор
    if (deleted.some(node => node.id === selectedNode?.id)) {
      setSelectedNode(null);
    }
  }, [selectedNode]);

  // Сохранение схемы
  const handleSave = async () => {
    setIsSaving(true);

    // Убираем поле selected перед сохранением
    const cleanNodes = nodes.map(({ data, ...node }) => ({
      ...node,
      data: {
        ...data,
        selected: undefined,
      },
    }));

    const result = await updateSchema(id, {
      name: schemaName,
      description: currentSchema.description,
      definition: {
        nodes: cleanNodes,
        edges,
      },
    });

    setIsSaving(false);

    if (result.success) {
      alert('✅ Схема сохранена!');
    } else {
      alert('❌ Ошибка сохранения');
    }
  };

  return (
    <div className="editor-container">
      <header className="editor-header">
        <button onClick={() => navigate('/')} className="btn-back">
          ← Назад
        </button>
        <input
          type="text"
          value={schemaName}
          onChange={(e) => setSchemaName(e.target.value)}
          className="schema-name-input"
          placeholder="Название схемы"
        />
        <div className="editor-actions">
          <button onClick={handleSave} className="btn-primary" disabled={isSaving}>
            {isSaving ? '💾 Сохранение...' : '💾 Сохранить'}
          </button>
          <button 
            onClick={() => setShowDebugModal(true)}
            className="btn-warning"
          >
            🐛 Отладка
          </button>
          <button 
            onClick={() => {
              console.log('🚀 Схема запущена!', { 
                schema_id: id, 
                nodes: nodes.length, 
                edges: edges.length 
              });
              alert('🚀 Схема отправлена на выполнение!\n(Функция выполнения будет реализована позже)');
            }} 
            className="btn-success"
          >
            ▶️ Запустить
          </button>
          <span className="editor-hint">
            💡 Удалить: выделить → Delete
          </span>
        </div>
      </header>

      <div className="editor-main">
        <NodesPalette />

        <div className="editor-canvas" ref={reactFlowWrapper}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            onNodesDelete={onNodesDelete}
            onInit={setReactFlowInstance}
            onDrop={onDrop}
            onDragOver={onDragOver}
            nodeTypes={nodeTypes}
            deleteKeyCode="Delete"
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>

        <NodeConfigPanel
          node={selectedNode}
          onUpdate={onNodeConfigUpdate}
          onClose={() => {
            setSelectedNode(null);
            setNodes((nds) =>
              nds.map((n) => ({
                ...n,
                data: { ...n.data, selected: false },
              }))
            );
          }}
        />
      </div>

      {showDebugModal && currentSchema && (
        <DebugModal 
          schema={currentSchema}
          onClose={() => setShowDebugModal(false)}
        />
      )}
    </div>
  );
}

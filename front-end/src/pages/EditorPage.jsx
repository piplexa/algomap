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
import ConditionNode from '../components/ConditionNode';
import NodesPalette from '../components/NodesPalette';
import NodeConfigPanel from '../components/NodeConfigPanel';
import DebugModal from '../components/DebugModal';
import { useSchemasStore } from '../store/schemasStore';
import { NODE_DEFINITIONS, NODE_TYPES } from '../utils/nodeTypes';
import { executionsAPI } from '../api/client';
import '../styles/Editor.css';

// Регистрируем компоненты для каждого типа нод
const nodeTypes = {
  [NODE_TYPES.START]: CustomNode,
  [NODE_TYPES.END]: CustomNode,
  [NODE_TYPES.LOG]: CustomNode,
  [NODE_TYPES.HTTP_REQUEST]: CustomNode,
  [NODE_TYPES.CONDITION]: ConditionNode,  // Условие рендерится как ромб
  [NODE_TYPES.VARIABLE_SET]: CustomNode,
  [NODE_TYPES.SLEEP]: CustomNode,
  [NODE_TYPES.MATH]: CustomNode,
  [NODE_TYPES.RABBITMQ_PUBLISH]: CustomNode,
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
  const [schemaStatus, setSchemaStatus] = useState(2); // По умолчанию active
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [showDebugModal, setShowDebugModal] = useState(false);

  // Статусы схемы из справочника dict_schema_status
  const schemaStatuses = [
    { id: 1, name: 'draft', label: 'Черновик', description: 'схема в разработке' },
    { id: 2, name: 'active', label: 'Активна', description: 'схема работает' },
    { id: 3, name: 'archived', label: 'Архив', description: 'схема устарела' },
  ];

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
      setSchemaStatus(currentSchema.status || 2); // По умолчанию active, если статус не указан

      if (currentSchema.definition?.nodes) {
        const loadedNodes = currentSchema.definition.nodes.map((node) => ({
          ...node,
          // Используем реальный тип из data, или если его нет - из node.type
          type: node.data?.type || node.type,
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
        type: nodeType,  // Используем реальный тип ноды
        position,
        data: {
          type: nodeType,  // Сохраняем тип и в data для совместимости
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
    // Теперь node.type содержит реальный тип (start, log, condition и т.д.)
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
      status: schemaStatus,
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

  // Запуск схемы
  const handleRunSchema = async () => {
    setIsRunning(true);

    try {
      const response = await executionsAPI.create(id);
      const execution = response.data;

      console.log('🚀 Схема запущена!', {
        schema_id: id,
        execution_id: execution.id,
        nodes: nodes.length,
        edges: edges.length
      });

      alert(`✅ Схема отправлена на выполнение!\nExecution ID: ${execution.id}`);
    } catch (error) {
      console.error('Ошибка запуска схемы:', error);
      alert(`❌ Ошибка запуска схемы: ${error.response?.data?.error || error.message}`);
    } finally {
      setIsRunning(false);
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
        <select
          value={schemaStatus}
          onChange={(e) => setSchemaStatus(parseInt(e.target.value, 10))}
          className="schema-status-select"
          title="Статус схемы"
        >
          {schemaStatuses.map((status) => (
            <option key={status.id} value={status.id}>
              {status.label}
            </option>
          ))}
        </select>
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
            onClick={handleRunSchema}
            className="btn-success"
            disabled={isRunning}
          >
            {isRunning ? '⏳ Запуск...' : '▶️ Запустить'}
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

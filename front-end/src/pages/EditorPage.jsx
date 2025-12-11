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
import EdgeConfigPanel from '../components/EdgeConfigPanel';
import ExecutionPanel from '../components/ExecutionPanel';
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
  const [selectedEdge, setSelectedEdge] = useState(null); // Состояние для выбранной линии
  const [schemaName, setSchemaName] = useState('');
  const [schemaDescription, setSchemaDescription] = useState('');
  const [schemaStatus, setSchemaStatus] = useState(2); // По умолчанию active
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [highlightedNodeId, setHighlightedNodeId] = useState(null);

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
      setSchemaDescription(currentSchema.description || '');
      setSchemaStatus(currentSchema.status || 2); // По умолчанию active, если статус не указан

      if (currentSchema.definition?.nodes) {
        const loadedNodes = currentSchema.definition.nodes.map((node) => ({
          ...node,
          // Используем реальный тип из data, или если его нет - из node.type
          type: node.data?.type || node.type,
          data: {
            ...node.data,
            selected: false,
            highlighted: false,
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
        // Применяем улучшенные стили к загруженным линиям
        const enhancedEdges = currentSchema.definition.edges.map((edge) => ({
          ...edge,
          // Если тип не указан или это default, меняем на smoothstep
          type: edge.type && edge.type !== 'default' ? edge.type : 'smoothstep',
          style: {
            stroke: '#94a3b8',
            strokeWidth: 2,
            ...edge.style, // Сохраняем пользовательские стили, если есть
          },
          markerEnd: edge.markerEnd || {
            type: 'arrowclosed',
            color: '#94a3b8',
          },
        }));
        setEdges(enhancedEdges);
      }
    }
  }, [currentSchema, setNodes, setEdges]);

  // Подсветка ноды при отладке/просмотре истории
  const handleNodeHighlight = useCallback(
    (nodeId) => {
      setHighlightedNodeId(nodeId);
      setNodes((nds) =>
        nds.map((n) => ({
          ...n,
          data: { ...n.data, highlighted: n.id === nodeId },
        }))
      );
      setEdges((eds) =>
        eds.map((e) => ({
          ...e,
          animated: e.target === nodeId,
          style: {
            ...e.style,
            stroke: e.target === nodeId ? '#f59e0b' : undefined,
            strokeWidth: e.target === nodeId ? 3 : undefined,
          },
        }))
      );
    },
    [setNodes, setEdges]
  );

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
  // Настраиваем дефолтные параметры для новых линий
  const onConnect = useCallback(
    (params) => {
      // Добавляем тип smoothstep для автоматического обхода блоков
      // и кастомные стили для лучшей читаемости
      const newEdge = {
        ...params,
        type: 'smoothstep', // Используем smoothstep вместо default
        animated: false, // По умолчанию без анимации
        style: {
          stroke: '#94a3b8', // Серо-синий цвет линии
          strokeWidth: 2, // Толщина линии
        },
        markerEnd: {
          type: 'arrowclosed', // Стрелка на конце
          color: '#94a3b8',
        },
      };
      setEdges((eds) => addEdge(newEdge, eds));
    },
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
      setSelectedEdge(null); // Сбрасываем выбор линии при выборе ноды
    },
    [setNodes]
  );

  // Выбор линии соединения
  const onEdgeClick = useCallback(
    (event, edge) => {
      event.stopPropagation();
      setSelectedEdge(edge);
      setSelectedNode(null); // Сбрасываем выбор ноды при выборе линии

      // Визуально выделяем выбранную линию
      setEdges((eds) =>
        eds.map((e) => ({
          ...e,
          selected: e.id === edge.id,
        }))
      );
    },
    [setEdges]
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

  // Обновление настроек линии (тип, анимация и т.д.)
  const onEdgeConfigUpdate = useCallback(
    (edgeId, updates) => {
      setEdges((eds) =>
        eds.map((edge) => {
          if (edge.id === edgeId) {
            return {
              ...edge,
              ...updates,
            };
          }
          return edge;
        })
      );

      // Обновляем выбранную линию в состоянии
      setSelectedEdge((current) =>
        current?.id === edgeId ? { ...current, ...updates } : current
      );
    },
    [setEdges]
  );

  // Удаление ноды
  const onNodesDelete = useCallback((deleted) => {
    // Если удалили выбранную ноду, сбрасываем выбор
    if (deleted.some(node => node.id === selectedNode?.id)) {
      setSelectedNode(null);
    }
  }, [selectedNode]);

  // Удаление линий
  const onEdgesDelete = useCallback((deleted) => {
    // Если удалили выбранную линию, сбрасываем выбор
    if (deleted.some(edge => edge.id === selectedEdge?.id)) {
      setSelectedEdge(null);
    }
  }, [selectedEdge]);

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
      description: schemaDescription,
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
        <div className="schema-info">
          <input
            type="text"
            value={schemaName}
            onChange={(e) => setSchemaName(e.target.value)}
            className="schema-name-input"
            placeholder="Название схемы"
          />
          <textarea
            value={schemaDescription}
            onChange={(e) => setSchemaDescription(e.target.value)}
            className="schema-description-input"
            placeholder="Описание схемы (необязательно)"
            rows={1}
          />
        </div>
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
            onEdgeClick={onEdgeClick}
            onNodesDelete={onNodesDelete}
            onEdgesDelete={onEdgesDelete}
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

        {/* Показываем либо панель настройки линий, либо панель настройки нод */}
        {selectedEdge ? (
          <EdgeConfigPanel
            edge={selectedEdge}
            onUpdate={onEdgeConfigUpdate}
            onClose={() => {
              setSelectedEdge(null);
              setEdges((eds) =>
                eds.map((e) => ({
                  ...e,
                  selected: false,
                }))
              );
            }}
          />
        ) : (
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
        )}

        <ExecutionPanel
          schemaId={id}
          isOpen={isPanelOpen}
          onToggle={() => setIsPanelOpen(!isPanelOpen)}
          onNodeHighlight={handleNodeHighlight}
        />
      </div>
    </div>
  );
}

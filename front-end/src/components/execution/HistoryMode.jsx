import { useState, useEffect } from 'react';
import ExecutionItem from './ExecutionItem';
import ExecutionDetails from './ExecutionDetails';
import ResizablePanels from './ResizablePanels';

export default function HistoryMode({ schemaId, onNodeHighlight }) {
  const [executions, setExecutions] = useState([]);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [executionDetails, setExecutionDetails] = useState(null);
  const [selectedStep, setSelectedStep] = useState(null);
  const [loading, setLoading] = useState(false);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  // Загрузка списка запусков
  useEffect(() => {
    if (schemaId) {
      loadExecutions();
    }
  }, [schemaId]);

  const loadExecutions = async () => {
    setLoading(true);
    try {
      // TODO: Заменить на реальный API вызов
      // const response = await fetch(`/api/executions?schema_id=${schemaId}&limit=100&offset=${offset}`);
      // const data = await response.json();

      // Пока используем заглушку
      const mockData = [
        {
          id: 'exec-1',
          schema_id: schemaId,
          status: 'completed',
          started_at: '2025-12-05T14:23:45Z',
          finished_at: '2025-12-05T14:23:50Z',
          steps_count: 5,
          duration: 2300,
        },
        {
          id: 'exec-2',
          schema_id: schemaId,
          status: 'error',
          started_at: '2025-12-04T12:10:30Z',
          finished_at: '2025-12-04T12:10:33Z',
          steps_count: 3,
          duration: 1200,
        },
      ];

      setExecutions(mockData);
      setHasMore(false);
    } catch (error) {
      console.error('Ошибка загрузки запусков:', error);
    } finally {
      setLoading(false);
    }
  };

  // Загрузка подробностей запуска
  const handleSelectExecution = async (executionId) => {
    setSelectedExecution(executionId);
    setSelectedStep(null);
    setLoading(true);

    try {
      // TODO: Заменить на реальный API вызов
      // const response = await fetch(`/api/executions/${executionId}`);
      // const data = await response.json();

      // Пока используем заглушку
      const mockDetails = {
        id: executionId,
        schema_id: schemaId,
        status: 'completed',
        started_at: '2025-12-05T14:23:45Z',
        finished_at: '2025-12-05T14:23:50Z',
        steps: [
          {
            id: 'step-1',
            node_id: 'start_1',
            node_type: 'start',
            status: 'success',
            started_at: '2025-12-05T14:23:45Z',
            finished_at: '2025-12-05T14:23:45Z',
            context_snapshot: {},
          },
          {
            id: 'step-2',
            node_id: 'variable_set_2',
            node_type: 'variable_set',
            status: 'success',
            started_at: '2025-12-05T14:23:46Z',
            finished_at: '2025-12-05T14:23:46Z',
            context_snapshot: { n: 5 },
          },
          {
            id: 'step-3',
            node_id: 'math_3',
            node_type: 'math',
            status: 'success',
            started_at: '2025-12-05T14:23:47Z',
            finished_at: '2025-12-05T14:23:47Z',
            context_snapshot: { n: 5, p: 0 },
          },
          {
            id: 'step-4',
            node_id: 'log_4',
            node_type: 'log',
            status: 'success',
            started_at: '2025-12-05T14:23:48Z',
            finished_at: '2025-12-05T14:23:48Z',
            context_snapshot: { n: 5, p: 0 },
            log_message: 'Результат: 0',
          },
          {
            id: 'step-5',
            node_id: 'end_5',
            node_type: 'end',
            status: 'success',
            started_at: '2025-12-05T14:23:50Z',
            finished_at: '2025-12-05T14:23:50Z',
            context_snapshot: { n: 5, p: 0 },
          },
        ],
      };

      setExecutionDetails(mockDetails);
    } catch (error) {
      console.error('Ошибка загрузки деталей запуска:', error);
    } finally {
      setLoading(false);
    }
  };

  // Выбор шага
  const handleSelectStep = (stepId) => {
    setSelectedStep(stepId);
    const step = executionDetails.steps.find((s) => s.id === stepId);
    if (step && onNodeHighlight) {
      onNodeHighlight(step.node_id);
    }
  };

  // Загрузка следующих 100 запусков при прокрутке до конца
  const handleLoadMore = () => {
    if (hasMore && !loading) {
      setOffset(offset + 100);
      loadExecutions();
    }
  };

  const leftPanel = (
    <div className="executions-list">
      <h4>Запуски</h4>
      <div className="executions-scroll" onScroll={(e) => {
        const bottom = e.target.scrollHeight - e.target.scrollTop === e.target.clientHeight;
        if (bottom) {
          handleLoadMore();
        }
      }}>
        {executions.length === 0 && !loading && (
          <div className="empty-state">Нет выполнений</div>
        )}
        {executions.map((exec) => (
          <ExecutionItem
            key={exec.id}
            execution={exec}
            isSelected={selectedExecution === exec.id}
            onClick={() => handleSelectExecution(exec.id)}
          />
        ))}
        {loading && <div className="loading-state">Загрузка...</div>}
      </div>
    </div>
  );

  const rightPanel = !selectedExecution ? (
    <div className="empty-state">Выберите запуск из списка</div>
  ) : (
    <ExecutionDetails
      steps={executionDetails?.steps || []}
      selectedStep={selectedStep}
      onSelectStep={handleSelectStep}
    />
  );

  return (
    <div className="history-mode">
      <ResizablePanels
        leftPanel={leftPanel}
        rightPanel={rightPanel}
        defaultLeftWidth={30}
      />
    </div>
  );
}

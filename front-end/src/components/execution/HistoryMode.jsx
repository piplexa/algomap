import { useState, useEffect } from 'react';
import ExecutionItem from './ExecutionItem';
import ExecutionDetails from './ExecutionDetails';
import ResizablePanels from './ResizablePanels';
import { executionsAPI } from '../../api/client';

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
      // Получаем список запусков из API
      const response = await executionsAPI.getList(schemaId, 100, offset);
      setExecutions(response.data);
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
      const response = await executionsAPI.getListSteps(executionId);
      // TODO: Ничего не понимаю в долбанном React, если передавать просто переменную response.data - не работает, но если передать объект с индексом steps - работает
      // Догадываюсь, что дело в этом: export default function ExecutionDetails({ steps, selectedStep, onSelectStep }) { (см. файл ExecutionDetails.jsx) вот совершенно не очевидно.
      let p = {
        steps: response.data
      };
      setExecutionDetails( p );
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

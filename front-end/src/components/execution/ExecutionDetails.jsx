import { useEffect } from 'react';
import StepsTab from './StepsTab';
import LogsTab from './LogsTab';
import ContextTab from './ContextTab';

export default function ExecutionDetails({ steps, selectedStep, onSelectStep }) {
  useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
        event.preventDefault();

        const currentIndex = steps.findIndex(step => step.id === selectedStep);

        if (event.key === 'ArrowDown') {
          // Переход на следующий шаг
          if (currentIndex < steps.length - 1) {
            onSelectStep(steps[currentIndex + 1].id);
          }
        } else if (event.key === 'ArrowUp') {
          // Переход на предыдущий шаг
          if (currentIndex > 0) {
            onSelectStep(steps[currentIndex - 1].id);
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [steps, selectedStep, onSelectStep]);

  return (
    <div className="execution-details-three-columns">
      {/* Колонка 1: Шаги (10%) */}
      <div className="steps-column">
        <h4>Шаги</h4>
        <div className="steps-column-content">
          <StepsTab
            steps={steps}
            selectedStep={selectedStep}
            onSelectStep={onSelectStep}
          />
        </div>
      </div>

      {/* Колонка 2: Контекст (45%) */}
      <div className="context-column">
        <h4>Контекст</h4>
        <div className="context-column-content">
          <ContextTab steps={steps} selectedStep={selectedStep} />
        </div>
      </div>

      {/* Колонка 3: Логи (45%) */}
      {
      /*
      <div className="logs-column">
        <h4>Логи</h4>
        <div className="logs-column-content">
          <LogsTab steps={steps} selectedStep={selectedStep} />
        </div>
      </div>
      */
      }
    </div>
  );
}

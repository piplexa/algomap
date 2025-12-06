# Задача: Панель истории выполнений и отладки

## Цель
Добавить выдвижную панель снизу редактора схем с двумя режимами:
1. **История** - просмотр выполненных запусков
2. **Отладка** - пошаговое выполнение схемы в реальном времени

## Визуальный дизайн

### Режим 1: История (пассивный просмотр)
```
┌─ 📊 История | 🐛 Отладка ─────────────────────────────────────┐
│                                                                │
│ ┌─ Запуски ────┐  ┌─ Шаги | Логи | Контекст ───────────────┐ │
│ │ 05.12 14:23  │  │ Просмотр выполненных шагов              │ │
│ │ ✓ Успех      │  │                                         │ │
│ │              │  │                                         │ │
│ │ 04.12 12:10  │  │                                         │ │
│ │ ✗ Ошибка     │  │                                         │ │
│ └──────────────┘  └─────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
```

### Режим 2: Отладка (активное выполнение)
```
┌─ 📊 История | 🐛 Отладка ─────────────────────────────────────┐
│ Выполнение: Шаг 3 из 5      ⏸️ Пауза | ⏭️ Шаг | ▶️ Продолжить │
│                                                                │
│ ┌─ Шаги ───────┐  ┌─ Шаги | Логи | Контекст ───────────────┐ │
│ │ ✓ Start      │  │ Текущий контекст в реальном времени     │ │
│ │ ✓ Set Var    │  │ {                                       │ │
│ │ ⚡ Math ←     │  │   "n": 5,                               │ │
│ │   Log        │  │   "p": 0                                │ │
│ │   End        │  │ }                                       │ │
│ └──────────────┘  └─────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
```

---

## Компонентная структура

### 1. ExecutionPanel.jsx (главный компонент)
**Расположение:** `front-end/src/components/ExecutionPanel.jsx`

**Props:**
- `schemaId` - ID текущей схемы
- `isOpen` - состояние панели (открыта/закрыта)
- `onToggle` - callback для открытия/закрытия

**State:**
- `mode` - 'history' | 'debug'
- `selectedExecution` - ID выбранного запуска
- `selectedStep` - ID выбранного шага
- `height` - высота панели (для resize)

**Структура:**
```jsx
<div className="execution-panel">
  {/* Заголовок с переключателем режимов */}
  <div className="panel-header">
    <button onClick={() => setMode('history')}>📊 История</button>
    <button onClick={() => setMode('debug')}>🐛 Отладка</button>
    <button className="panel-toggle" onClick={onToggle}>▼</button>
  </div>

  {/* Контент в зависимости от режима */}
  <div className="panel-content">
    {mode === 'history' ? (
      <HistoryMode schemaId={schemaId} />
    ) : (
      <DebugMode schemaId={schemaId} />
    )}
  </div>
</div>
```

---

### 2. HistoryMode.jsx
**Расположение:** `front-end/src/components/execution/HistoryMode.jsx`

**Функционал:**
- Загрузка списка запусков из `GET /api/executions?schema_id={schemaId}`
- Левая колонка: список запусков
- Правая колонка: три вкладки (Шаги | Логи | Контекст)
- При клике на запуск → загружаются шаги
- При клике на шаг → подсветка нода на схеме + контекст этого шага

**Структура:**
```jsx
<div className="history-mode">
  {/* Левая колонка: список запусков */}
  <div className="executions-list">
    <h4>Запуски</h4>
    <div className="executions-scroll">
      {executions.map(exec => (
        <ExecutionItem 
          key={exec.id}
          execution={exec}
          isSelected={selectedExecution === exec.id}
          onClick={() => handleSelectExecution(exec.id)}
        />
      ))}
    </div>
  </div>

  {/* Правая колонка: вкладки */}
  <div className="execution-details">
    <div className="tabs">
      <button onClick={() => setTab('steps')}>Шаги</button>
      <button onClick={() => setTab('logs')}>Логи</button>
      <button onClick={() => setTab('context')}>Контекст</button>
    </div>

    <div className="tab-content">
      {tab === 'steps' && <StepsTab executionId={selectedExecution} />}
      {tab === 'logs' && <LogsTab executionId={selectedExecution} />}
      {tab === 'context' && <ContextTab executionId={selectedExecution} stepId={selectedStep} />}
    </div>
  </div>
</div>
```

**API запросы:**
```javascript
// Список запусков
GET /api/executions?schema_id={schemaId}

// Детали запуска со всеми шагами
GET /api/executions/{executionId}

// Ответ:
{
  "id": "uuid",
  "schema_id": "uuid",
  "status": "completed",
  "started_at": "2025-12-05T14:23:45Z",
  "finished_at": "2025-12-05T14:23:50Z",
  "steps": [
    {
      "id": "uuid",
      "node_id": "start_1",
      "node_type": "start",
      "status": "success",
      "started_at": "...",
      "finished_at": "...",
      "context_snapshot": { "n": 5, "p": 0 }
    },
    // ...
  ]
}
```

---

### 3. DebugMode.jsx
**Расположение:** `front-end/src/components/execution/DebugMode.jsx`

**Функционал:**
- Запуск схемы в режиме отладки
- Управление: Пауза | Шаг | Продолжить
- Левая колонка: текущий прогресс выполнения
- Правая колонка: контекст в реальном времени

**Структура:**
```jsx
<div className="debug-mode">
  {/* Статус и кнопки управления */}
  <div className="debug-controls">
    <div className="debug-status">
      Выполнение: Шаг {currentStep} из {totalSteps}
    </div>
    <div className="debug-buttons">
      <button onClick={handlePause}>⏸️ Пауза</button>
      <button onClick={handleNextStep}>⏭️ Шаг</button>
      <button onClick={handleContinue}>▶️ Продолжить</button>
    </div>
  </div>

  {/* Основной контент */}
  <div className="debug-content">
    {/* Левая колонка: прогресс */}
    <div className="debug-progress">
      <h4>Шаги</h4>
      {steps.map((step, idx) => (
        <div 
          key={idx} 
          className={`step-item ${getStepStatus(idx)}`}
        >
          {getStepIcon(idx)} {step.label}
        </div>
      ))}
    </div>

    {/* Правая колонка: вкладки */}
    <div className="debug-details">
      <div className="tabs">
        <button onClick={() => setTab('steps')}>Шаги</button>
        <button onClick={() => setTab('logs')}>Логи</button>
        <button onClick={() => setTab('context')}>Контекст</button>
      </div>

      <div className="tab-content">
        {tab === 'context' && (
          <pre className="context-json">
            {JSON.stringify(currentContext, null, 2)}
          </pre>
        )}
      </div>
    </div>
  </div>
</div>
```

**Логика отладки:**
1. При нажатии "Запустить в режиме отладки" → `POST /api/executions/debug/{schemaId}`
2. Backend создает execution и останавливается на первом шаге
3. Frontend получает execution_id
4. При нажатии "Шаг" → `POST /api/executions/{executionId}/step`
5. Backend выполняет один шаг, возвращает обновленное состояние
6. Frontend обновляет UI + подсвечивает текущий нод на схеме

**API для отладки:**
```javascript
// Запуск в режиме отладки
POST /api/executions/debug/{schemaId}
Response: { execution_id: "uuid", current_step: {...} }

// Выполнить один шаг
POST /api/executions/{executionId}/step
Response: { 
  current_step: {...},
  context: {...},
  is_finished: false
}

// Продолжить выполнение
POST /api/executions/{executionId}/continue

// Пауза
POST /api/executions/{executionId}/pause
```

---

### 4. Вспомогательные компоненты

#### ExecutionItem.jsx
Одна строка в списке запусков
```jsx
<div className="execution-item">
  <div className="execution-time">05.12.2025 14:23:45</div>
  <div className="execution-status">
    {status === 'completed' ? '✓ Успех' : '✗ Ошибка'}
  </div>
  <div className="execution-duration">5 шагов | 2.3с</div>
</div>
```

#### StepsTab.jsx
Список шагов выполнения
```jsx
<div className="steps-list">
  {steps.map(step => (
    <div 
      key={step.id} 
      className="step-row"
      onClick={() => onSelectStep(step.id)}
    >
      <span className="step-icon">{getNodeIcon(step.node_type)}</span>
      <span className="step-name">{step.node_id}</span>
      <span className="step-status">{step.status}</span>
      <span className="step-time">{formatDuration(step)}</span>
    </div>
  ))}
</div>
```

#### LogsTab.jsx
Логи от Log-нодов
```jsx
<div className="logs-list">
  {logs.map(log => (
    <div key={log.id} className="log-entry">
      <span className="log-time">{formatTime(log.timestamp)}</span>
      <span className="log-message">{log.message}</span>
    </div>
  ))}
</div>
```

#### ContextTab.jsx
JSON-контекст с подсветкой синтаксиса
```jsx
<pre className="context-viewer">
  <code>{JSON.stringify(context, null, 2)}</code>
</pre>
```

---

## Стили (ExecutionPanel.css)

### Общие стили панели
```css
/* Выдвижная панель снизу */
.execution-panel {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 2px solid var(--gray-300);
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  z-index: 100;
  transition: height 0.3s ease;
}

.execution-panel.closed {
  height: 40px;
}

.execution-panel.open {
  height: 400px;
  min-height: 300px;
  max-height: 70vh;
  resize: vertical;
  overflow: hidden;
}

/* Заголовок панели */
.panel-header {
  height: 40px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  border-bottom: 1px solid var(--gray-200);
  background: var(--gray-50);
}

.panel-header button {
  padding: 6px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
  transition: background 0.2s;
}

.panel-header button.active {
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.panel-toggle {
  margin-left: auto;
  font-size: 16px;
}

/* Контент панели */
.panel-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}
```

### Режим истории
```css
.history-mode {
  display: flex;
  width: 100%;
  height: 100%;
}

/* Левая колонка: список запусков */
.executions-list {
  width: 30%;
  min-width: 250px;
  border-right: 1px solid var(--gray-200);
  display: flex;
  flex-direction: column;
  background: var(--gray-50);
}

.executions-list h4 {
  padding: 12px 16px;
  margin: 0;
  font-size: 13px;
  color: var(--gray-600);
  background: var(--gray-100);
  border-bottom: 1px solid var(--gray-200);
}

.executions-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.execution-item {
  padding: 12px;
  margin-bottom: 4px;
  background: white;
  border: 1px solid var(--gray-200);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.execution-item:hover {
  border-color: var(--primary);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.execution-item.selected {
  border-color: var(--primary);
  background: var(--primary-light);
}

.execution-time {
  font-size: 12px;
  color: var(--gray-600);
  margin-bottom: 4px;
}

.execution-status {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
}

.execution-status.success {
  color: var(--success);
}

.execution-status.error {
  color: var(--danger);
}

.execution-duration {
  font-size: 11px;
  color: var(--gray-500);
}

/* Правая колонка: детали */
.execution-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tabs {
  display: flex;
  gap: 4px;
  padding: 8px 16px;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-200);
}

.tabs button {
  padding: 6px 16px;
  border: none;
  background: transparent;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.tabs button.active {
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.tab-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
```

### Вкладка Шаги
```css
.steps-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.step-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--gray-50);
  border: 1px solid var(--gray-200);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.step-row:hover {
  background: white;
  border-color: var(--primary);
}

.step-row.selected {
  background: var(--primary-light);
  border-color: var(--primary);
}

.step-icon {
  font-size: 18px;
}

.step-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-700);
}

.step-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 12px;
}

.step-status.success {
  background: var(--success-light);
  color: var(--success-dark);
}

.step-status.error {
  background: var(--danger-light);
  color: var(--danger-dark);
}

.step-time {
  font-size: 11px;
  color: var(--gray-500);
}
```

### Вкладка Логи
```css
.logs-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-family: monospace;
  font-size: 12px;
}

.log-entry {
  display: flex;
  gap: 12px;
  padding: 8px 12px;
  background: var(--gray-50);
  border-left: 3px solid var(--primary);
  border-radius: 4px;
}

.log-time {
  color: var(--gray-500);
  min-width: 80px;
}

.log-message {
  flex: 1;
  color: var(--gray-800);
  word-break: break-word;
}
```

### Вкладка Контекст
```css
.context-viewer {
  margin: 0;
  padding: 16px;
  background: var(--gray-900);
  color: var(--gray-100);
  border-radius: 6px;
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
  overflow-x: auto;
}

.context-viewer code {
  color: inherit;
}
```

### Режим отладки
```css
.debug-mode {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
}

/* Управление отладкой */
.debug-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--warning-light);
  border-bottom: 1px solid var(--warning);
}

.debug-status {
  font-size: 14px;
  font-weight: 500;
  color: var(--gray-800);
}

.debug-buttons {
  display: flex;
  gap: 8px;
}

.debug-buttons button {
  padding: 6px 12px;
  border: 1px solid var(--gray-300);
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.debug-buttons button:hover {
  background: var(--gray-50);
  border-color: var(--primary);
}

/* Контент отладки */
.debug-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.debug-progress {
  width: 30%;
  min-width: 250px;
  border-right: 1px solid var(--gray-200);
  display: flex;
  flex-direction: column;
  background: var(--gray-50);
  overflow-y: auto;
  padding: 12px;
}

.debug-progress h4 {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: var(--gray-600);
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 4px;
  border-radius: 6px;
  font-size: 13px;
  transition: all 0.2s;
}

.step-item.completed {
  background: var(--success-light);
  color: var(--success-dark);
}

.step-item.current {
  background: var(--warning);
  color: white;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.step-item.pending {
  color: var(--gray-500);
  background: transparent;
}

.debug-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}
```

---

## Интеграция с редактором

### В EditorPage.jsx добавить:

```jsx
import ExecutionPanel from '../components/ExecutionPanel';

// State
const [isPanelOpen, setIsPanelOpen] = useState(false);

// В JSX после .editor-canvas
<ExecutionPanel 
  schemaId={id}
  isOpen={isPanelOpen}
  onToggle={() => setIsPanelOpen(!isPanelOpen)}
/>
```

### Подсветка нодов на схеме

Когда выбран шаг из истории или текущий шаг в отладке:

```jsx
// При выборе шага обновляем ноды
const highlightNode = (nodeId) => {
  setNodes(nodes.map(node => ({
    ...node,
    data: {
      ...node.data,
      highlighted: node.id === nodeId
    }
  })));
};

// В CustomNode.jsx добавить стиль
<div className={`custom-node ${data.highlighted ? 'highlighted' : ''}`}>
  ...
</div>
```

```css
.custom-node.highlighted {
  border: 3px solid var(--warning);
  box-shadow: 0 0 0 4px rgba(251, 191, 36, 0.2);
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}
```

---

## Чеклист реализации

### Фаза 1: Структура и UI
- [ ] Создать `ExecutionPanel.jsx` с базовой структурой
- [ ] Добавить стили `ExecutionPanel.css`
- [ ] Реализовать открытие/закрытие панели (▼/▲)
- [ ] Добавить переключатель режимов (История/Отладка)
- [ ] Интегрировать в `EditorPage.jsx`

### Фаза 2: Режим История
- [ ] Создать `HistoryMode.jsx`
- [ ] Реализовать `ExecutionItem.jsx`
- [ ] Подключить API `GET /api/executions`
- [ ] Реализовать `StepsTab.jsx`
- [ ] Реализовать `LogsTab.jsx`
- [ ] Реализовать `ContextTab.jsx`
- [ ] Добавить подсветку нодов при выборе шага

### Фаза 3: Режим Отладка
- [ ] Создать `DebugMode.jsx`
- [ ] Реализовать кнопки управления (Пауза/Шаг/Продолжить)
- [ ] Подключить API отладки
- [ ] Синхронизация состояния с backend через polling/SSE
- [ ] Подсветка текущего нода в реальном времени
- [ ] Обработка ошибок при выполнении

### Фаза 4: Полировка
- [ ] Адаптивная высота панели (resize)
- [ ] Keyboard shortcuts (Space = следующий шаг, Esc = закрыть)
- [ ] Loading states и скелетоны
- [ ] Пустые состояния ("Нет выполнений", "Выберите запуск")
- [ ] Тестирование на разных размерах экрана

---

## Backend Requirements

Для работы панели потребуются следующие эндпоинты:

### История
```
GET /api/executions?schema_id={id}&limit=20&offset=0
GET /api/executions/{executionId}
```

### Отладка
```
POST /api/executions/debug/{schemaId} - начать отладку (возвратит executionId)
POST /api/executions/{executionId}/step - перейти на один шаг. Тоесть выполнить следующий  шаг из executions
```

Эти методы оставить на потом, потому как не совсем понятно как с ними работать.
```
POST /api/executions/{executionId}/pause
POST /api/executions/{executionId}/continue
GET  /api/executions/{executionId}/status (polling для обновления UI)
```

---

## Примечания

- Старый `DebugModal.jsx` можно удалить
- Цвета можно взять из существующих CSS переменных
- Иконки: использовать эмодзи или добавить библиотеку типа `lucide-react`
- Для JSON-подсветки можно использовать `react-json-view` или просто `<pre>`

---

## Вопросы для уточнения

1. Нужна ли пагинация в списке запусков или просто scroll? - Просто скрол. Сейчас не делать.
2. WebSocket/SSE для реального времени в отладке или polling? - Ни то ни другое. Отладка будет в виде запроса на каждый шаг. Пользователь просто будет нажимать каждый раз на кнопку.
3. Сохранять ли высоту панели в localStorage? - Да, можно. 
4. Максимум сколько запусков показывать в истории? Последние 100. А если дойдет до конца и то сделать запрос на с 101 по 200. потом с 201 по 300 и т.д.

---

**Удачи! 🚀**

// Типы нод из ТЗ: docs/TZ_Node_Types.md

export const NODE_TYPES = {
  START: 'start',
  END: 'end',
  LOG: 'log',
  HTTP_REQUEST: 'http_request',
  CONDITION: 'condition',
  VARIABLE_SET: 'variable_set',
  SLEEP: 'sleep',
  MATH: 'math',
  RABBITMQ_PUBLISH: 'rabbitmq_publish',
};

export const NODE_CATEGORIES = {
  FLOW: 'Управление потоком',
  DATA: 'Данные',
  EXTERNAL: 'Внешние вызовы',
  LOGIC: 'Логика',
  TIME: 'Время',
};

export const NODE_DEFINITIONS = {
  [NODE_TYPES.START]: {
    type: NODE_TYPES.START,
    label: 'Start',
    category: NODE_CATEGORIES.FLOW,
    icon: '▶️',
    color: '#10b981',
    outputs: 1,
    config: {},
  },
  [NODE_TYPES.END]: {
    type: NODE_TYPES.END,
    label: 'End',
    category: NODE_CATEGORIES.FLOW,
    icon: '⏹️',
    color: '#ef4444',
    outputs: 0,
    config: {
      success: true,
      message: 'Execution completed',
    },
  },
  [NODE_TYPES.LOG]: {
    type: NODE_TYPES.LOG,
    label: 'Log',
    category: NODE_CATEGORIES.LOGIC,
    icon: '📝',
    color: '#8b5cf6',
    outputs: 1,
    config: {
      level: 'info',
      message: '',
    },
  },
  [NODE_TYPES.HTTP_REQUEST]: {
    type: NODE_TYPES.HTTP_REQUEST,
    label: 'HTTP Request',
    category: NODE_CATEGORIES.EXTERNAL,
    icon: '🌐',
    color: '#3b82f6',
    outputs: 2, // success, error
    config: {
      method: 'GET',
      url: '',
      headers: {},
      body: {},
      timeout: 30,
      retry: {
        enabled: false,
        max_attempts: 3,
        delay: 5,
      },
    },
  },
  [NODE_TYPES.CONDITION]: {
    type: NODE_TYPES.CONDITION,
    label: 'Condition',
    category: NODE_CATEGORIES.FLOW,
    icon: '🔀',
    color: '#f59e0b',
    outputs: 2, // true, false
    config: {
      expression: '',
    },
  },
  [NODE_TYPES.VARIABLE_SET]: {
    type: NODE_TYPES.VARIABLE_SET,
    label: 'Set Variable',
    category: NODE_CATEGORIES.DATA,
    icon: '💾',
    color: '#06b6d4',
    outputs: 1,
    config: {
      variable: '',
      value: '',
    },
  },
  [NODE_TYPES.SLEEP]: {
    type: NODE_TYPES.SLEEP,
    label: 'Sleep',
    category: NODE_CATEGORIES.TIME,
    icon: '⏰',
    color: '#ec4899',
    outputs: 1,
    config: {
      duration: 60,
      unit: 'seconds',
    },
  },
  [NODE_TYPES.MATH]: {
    type: NODE_TYPES.MATH,
    label: 'Math',
    category: NODE_CATEGORIES.DATA,
    icon: '🔢',
    color: '#14b8a6',
    outputs: 1,
    config: {
      operation: 'add',
      operand1: '',
      operand2: '',
      result_variable: '',
    },
  },
  //// TODO: Работу с брокером очередей перенесена в следующие версии
  // [NODE_TYPES.RABBITMQ_PUBLISH]: {
  //   type: NODE_TYPES.RABBITMQ_PUBLISH,
  //   label: 'RabbitMQ',
  //   category: NODE_CATEGORIES.EXTERNAL,
  //   icon: '🐰',
  //   color: '#f97316',
  //   outputs: 2, // success, error
  //   config: {
  //     queue: '',
  //     exchange: '',
  //     message: {},
  //   },
  // },
};

// Группировка по категориям для панели инструментов
export const getNodesByCategory = () => {
  const grouped = {};
  
  Object.values(NODE_DEFINITIONS).forEach((node) => {
    if (!grouped[node.category]) {
      grouped[node.category] = [];
    }
    grouped[node.category].push(node);
  });
  
  return grouped;
};

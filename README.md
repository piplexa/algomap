# AlgoMap

Умение программировать — это не про знание Python или JavaScript.  
Это про умение **разложить задачу на шаги** и **построить логику**.

AlgoMap возвращает программирование к его сути: **алгоритмическому мышлению**.  
Блок-схемы — это то, как реально думает программист, до того как начнёт печатать код.

### Что нужно, чтобы быть программистом?

**Два простых понятия:**
1. **Как работает процессор** — минимум: 2 регистра и 2 команды, чтобы сложить два числа
2. **Что такое алгоритм** — последовательность шагов для решения задачи

Всё остальное — синтаксис, библиотеки, фреймворки — это детали, которые меняются каждые 5 лет.

В эпоху, когда ChatGPT пишет код за секунды, важно понимать:
- **Что** должна делать программа
- **Почему** она работает именно так  
- **Как** разбить проблему на части

Это навык, который нельзя сгенерировать. И именно этому учит AlgoMap.

[Проверить и попробовать](https://algo-map.ru/login)

## Описание

AlgoMap позволяет создавать и выполнять схемы автоматизации через интуитивный визуальный интерфейс. Соединяйте блоки (ноды) в логические цепочки и запускайте их через API, webhook или вручную.

## Возможности

- 🎨 Визуальный редактор схем (drag & drop)
- 🔄 Асинхронное выполнение через RabbitMQ
- ⏱️ Отложенные задачи (sleep/delay) (требуется at сервер)
- 🌐 HTTP запросы к внешним API
- 🔀 Условная логика (if/else, switch)
- 📊 Управление переменными
- 🪝 Webhook триггеры
- ⏸️ Pause/Resume выполнения

## Архитектура

```
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│   Frontend  │─────▶│   Backend API │─────▶│  PostgreSQL │
│   (React)   │       │   (Go REST)  │       │             │
└─────────────┘       └──────────────┘       └─────────────┘
                            │
                            ▼
                     ┌──────────────┐       ┌─────────────┐
                     │  RabbitMQ    │◀────▶│ Backend Core│
                     │   (Queue)    │       │  (Workers)  │
                     └──────────────┘       └─────────────┘
                                                  │
                                                  ▼
                                           ┌─────────────┐
                                           │  at Server  │
                                           │  (Delayed)  │
                                           └─────────────┘
```

## Технологический стек

### Frontend
- React 18+
- React Flow (визуальный редактор)
- TypeScript
- Axios

### Backend
- Go 1.21+
- PostgreSQL 16
- RabbitMQ 3
- [github.com/piplexa/at](https://github.com/piplexa/at) (отложенные задачи)

### Infrastructure
- Docker & Docker Compose
- Nginx (прокси в продакшене)

## Структура проекта

```
algomap/
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go          # REST API сервер
│   │   └── worker/
│   │       └── main.go          # Execution worker
│   │
│   ├── internal/
│   │   ├── domain/              # Модели данных
│   │   │   ├── schema.go
│   │   │   ├── execution.go
│   │   │   └── node.go
│   │   │
│   │   ├── handlers/            # HTTP handlers (только для API)
│   │   │   ├── schemas.go
│   │   │   ├── executions.go
│   │   │   └── webhook.go
│   │   │
│   │   ├── repository/          # Работа с БД
│   │   │   ├── postgres.go
│   │   │   ├── schemas.go
│   │   │   └── executions.go
│   │   │
│   │   ├── executor/            # Движок выполнения (только для worker)
│   │   │   ├── engine.go
│   │   │   ├── context.go
│   │   │   └── state.go
│   │   │
│   │   └── nodes/               # Обработчики типов нод (для worker)
│   │       ├── handler.go       # интерфейс NodeHandler
│   │       ├── start.go
│   │       ├── end.go
│   │       ├── http.go
│   │       ├── condition.go
│   │       ├── sleep.go
│   │       └── log.go
│   │
│   ├── pkg/                     # Публичные переиспользуемые пакеты
│   │   ├── logger/
│   │   │   └── logger.go
│   │   └── config/
│   │       └── config.go
│   ├── sql/
│   │   └── schema.sql
│   │
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   ├── public/
│   └── package.json
│
├── migrations/
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
│
├── docs/
│   └── (ТЗ файлы)
│
├── docker-compose.yml
├── .gitignore
└── README.md
```

## Быстрый старт

### Требования

- Docker & Docker Compose
- Go 1.21+ (для разработки)
- Node.js 18+ (для разработки)
- AT сервер (для отложенных задач)

### Запуск локально

```bash
# Клонировать репозиторий
git clone https://github.com/piplexa/algomap.git
cd algomap

# Запустить все сервисы через Docker Compose
docker-compose up -d

# Frontend будет доступен на http://localhost:3000
# Backend API на http://localhost:8080
# RabbitMQ Management UI на http://localhost:15672 (guest/guest)
```

### Разработка

```bash
# Frontend
cd frontend
npm install
npm start

# Backend API
cd backend
go run cmd/api/main.go

# Backend Worker
cd backend
go run cmd/worker/main.go
```

## Типы нод

Поддерживаемые типы блоков (нод):

### Управление потоком
- **Start** - точка входа
- **End** - завершение схемы
- **Condition** - условие (if/else)
- **Sleep** - задержка выполнения

### Данные
- **Variable Set** - установить переменную
- **Math** - математические операции

### Внешние вызовы
- **HTTP Request** - запрос к API
- **RabbitMQ Publish** - отправка сообщения в очередь

### Логика
- **Log** - вывод в лог

### Заглушки (MVP)
- **SubSchema** - вызов другой схемы (будет реализовано позже)

Полное описание в [docs/TZ_Node_Types.md](docs/TZ_Node_Types.md)

## API Endpoints

```
GET    /api/schemas          - список схем
POST   /api/schemas          - создать схему
GET    /api/schemas/:id      - получить схему
PUT    /api/schemas/:id      - обновить схему
DELETE /api/schemas/:id      - удалить схему

POST   /api/executions                - запустить выполнение
GET    /api/executions/:id            - статус выполнения
POST   /api/executions/:id/pause      - пауза
POST   /api/executions/:id/resume     - продолжить
POST   /api/executions/:id/stop       - остановить

POST   /api/executions/id-execution/id-node/continue            - webhook триггер
```

Полное API в [docs/TZ_Backend_API.md](docs/TZ_Backend_API.md)

## Переменные окружения

### Backend
```env
DATABASE_URL=postgresql://user:pass@localhost:5432/algomap
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
LOG_LEVEL=debug
PORT=8080
```

### Frontend
```env
REACT_APP_API_URL=http://localhost:8080/api
```

## База данных

PostgreSQL схема включает:
- `schemas` - определения схем
- `executions` - история выполнений
- `execution_state` - текущее состояние выполнения
- `execution_steps` - детали каждого шага

Миграции в папке `migrations/`

## Roadmap

### MVP (v0.1)
- [x] Техническое задание
- [x] Базовая структура проекта
- [x] Визуальный редактор схем
- [x] Базовые типы нод (Start, End, HTTP, Condition, Log)
- [x] REST API для схем
- [x] Execution engine с RabbitMQ
- [x] Sleep/Delay через `at` сервер
- [x] Webhook триггеры (через executions/.../continue )

### v0.2
- [x] Http запросы не блокируют worker
- [x] Обход блоков линий связей в алгоритме
- [ ] Работа с массивами
- [x] Дополнительные типы нод (Math, Variable)
- [ ] Pause/Resume выполнения
- [x] История выполнений
- [x] UI для логов
- [ ] RabbitMQ Publish

### v0.3
- [x] Циклы (Loop)
- [ ] Вложенные схемы (SubSchema)
- [x] Аутентификация пользователей
- [ ] Мультитенантность

### Будущее
- [ ] Планировщик (cron-like)
- [ ] Визуализация выполнения в реальном времени
- [ ] Marketplace нод/схем
- [ ] Кастомные ноды (плагины)
- [ ] Аутентификация через google

## Контрибьюция

Проект находится в стадии активной разработки. Pull requests приветствуются!

## Лицензия

MIT

## Авторы

- [@piplexa](https://github.com/piplexa)

---

**Статус проекта:** 🚧 В разработке (MVP)
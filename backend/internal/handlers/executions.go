package handlers

// ExecutionHandler - HTTP handlers для управления выполнениями схем

// В первую очередь:
// POST   /api/executions                	- запустить схему (manual)
// GET    /api/executions/:id/steps      	- история шагов
// GET    /api/executions/:id            	- статус выполнения

// TODO: Реализовать endpoints:
// POST   /api/executions/:id/pause      	- пауза
// POST   /api/executions/:id/resume     	- продолжить
// POST   /api/executions/:id/stop       	- остановить
// GET    /api/executions/:id/logs       	- логи выполнения
// POST   /api/executions/:id/:id/continue  - начать выполнение с указанного узла схемы
// POST   /api/executions/:id/:id/one    	- выполнить только указанный узел схемы
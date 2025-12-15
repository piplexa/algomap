package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/piplexa/algomap/internal/executor"
)

const (
	QueueName    = "schema_execution_queue"
	ExchangeName = "schema_execution"
)

// Consumer обработчик сообщений из RabbitMQ
type Consumer struct {
	conn      *Connection
	logger    *zap.Logger
	engine    *executor.Engine
	publisher *Publisher
}

// NewConsumer создаёт новый consumer
func NewConsumer(amqpURL string, logger *zap.Logger, engine *executor.Engine) (*Consumer, error) {
	// Создаём connection
	conn, err := NewConnection(amqpURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Создаём publisher для автопубликации следующих нод
	publisher := NewPublisher(conn, logger)

	// Объявляем очередь
	if err := publisher.DeclareQueue(QueueName); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:      conn,
		logger:    logger,
		engine:    engine,
		publisher: publisher,
	}, nil
}

// Start начинает обработку сообщений
func (c *Consumer) Start(ctx context.Context) error {
	channel, err := c.conn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Устанавливаем prefetch count = 1 для равномерного распределения
	if err := channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := channel.Consume(
		QueueName,
		"",    // consumer tag
		true, // auto-ack (отключаем, будем ACK вручную)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("worker started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("worker shutting down...")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				c.logger.Warn("channel closed")
				return fmt.Errorf("channel closed")
			}

			go c.handleMessage(ctx, msg)
		}
	}
}

// handleMessage обрабатывает одно сообщение
func (c *Consumer) handleMessage(ctx context.Context, msg amqp091.Delivery) {
	c.logger.Info("Сообщение из брокера очередей: ", zap.String("body", string(msg.Body)))

	var execMsg executor.ExecutionMessage
	if err := json.Unmarshal(msg.Body, &execMsg); err != nil {
		c.logger.Error("failed to unmarshal message", zap.Error(err))
		msg.Nack(false, false) // Отклоняем, не возвращаем в очередь
		return
	}

	// Выполняем ноду через engine (теперь возвращает nextNodeID)
	nextNodeID, needContinue, err := c.engine.Execute(ctx, &execMsg)
	c.logger.Log(
		-2, 
		"После выполнения ноды.", 
		zap.Bool("Надо ли продолжать выполнение схемы: ", *needContinue),
	);
	if err != nil {
		c.logger.Error("failed to execute node",
			zap.String("execution_id", execMsg.ExecutionID),
			zap.String("node_id", execMsg.CurrentNodeID),
			zap.Error(err),
		)
		return
	}

	// Если есть следующая нода - публикуем новое сообщение
	// Нода может быть sleep, тогда не нужно отправлять сообщение в очередь - это определяет needContinue
	if nextNodeID != nil && *needContinue {
		newMsg := &executor.ExecutionMessage{
			ExecutionID:   execMsg.ExecutionID,
			SchemaID:      execMsg.SchemaID,
			CurrentNodeID: *nextNodeID,
			DebugMode:     execMsg.DebugMode,
		}

		if err := c.publisher.Publish(ctx, QueueName, newMsg); err != nil {
			c.logger.Error("failed to publish next message",
				zap.String("execution_id", execMsg.ExecutionID),
				zap.String("next_node_id", *nextNodeID),
				zap.Error(err),
			)
			// TODO: Решить что делать если не удалось опубликовать
			// Варианты: retry, DLQ, отметить execution как failed
		} 
	}
}

// Close закрывает соединение
func (c *Consumer) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

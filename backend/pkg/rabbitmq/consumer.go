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
	conn    *amqp091.Connection
	channel *amqp091.Channel
	logger  *zap.Logger
	engine  *executor.Engine
}

// NewConsumer создаёт новый consumer
func NewConsumer(amqpURL string, logger *zap.Logger, engine *executor.Engine) (*Consumer, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Устанавливаем prefetch count = 1 для равномерного распределения
	if err := channel.Qos(1, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	// Объявляем очередь (durable)
	_, err = channel.QueueDeclare(
		QueueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		logger:  logger,
		engine:  engine,
	}, nil
}

// Start начинает обработку сообщений
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		QueueName,
		"",    // consumer tag
		false, // auto-ack (отключаем, будем ACK вручную)
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

			c.handleMessage(ctx, msg)
		}
	}
}

// handleMessage обрабатывает одно сообщение
func (c *Consumer) handleMessage(ctx context.Context, msg amqp091.Delivery) {
	c.logger.Debug("received message", zap.String("body", string(msg.Body)))

	var execMsg executor.ExecutionMessage
	if err := json.Unmarshal(msg.Body, &execMsg); err != nil {
		c.logger.Error("failed to unmarshal message", zap.Error(err))
		msg.Nack(false, false) // Отклоняем, не возвращаем в очередь
		return
	}

	// Выполняем ноду
	if err := c.engine.Execute(ctx, &execMsg); err != nil {
		c.logger.Error("failed to execute node",
			zap.String("execution_id", execMsg.ExecutionID),
			zap.String("node_id", execMsg.CurrentNodeID),
			zap.Error(err),
		)
		// TODO: Решить, что делать с failed нодами
		// Пока просто ACK, чтобы не блокировать очередь
		msg.Ack(false)
		return
	}

	// TODO: Если нода успешна и есть следующая - опубликовать новое сообщение
	// Это должен делать worker после успешного выполнения

	// Подтверждаем обработку
	if err := msg.Ack(false); err != nil {
		c.logger.Error("failed to ack message", zap.Error(err))
	}
}

// Close закрывает соединение
func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

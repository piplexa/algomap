package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Publisher публикует сообщения в RabbitMQ
type Publisher struct {
	conn   *Connection
	logger *zap.Logger
}

// NewPublisher создаёт новый publisher
func NewPublisher(conn *Connection, logger *zap.Logger) *Publisher {
	return &Publisher{
		conn:   conn,
		logger: logger,
	}
}

// DeclareQueue объявляет очередь (если не существует)
func (p *Publisher) DeclareQueue(queueName string) error {
	channel, err := p.conn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable - очередь переживёт перезапуск RabbitMQ
		false,     // autoDelete - не удалять когда нет подписчиков
		false,     // exclusive - не эксклюзивная
		false,     // noWait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	p.logger.Info("Queue declared", zap.String("queue", queueName))
	return nil
}

// Publish публикует сообщение в очередь
func (p *Publisher) Publish(ctx context.Context, queueName string, message interface{}) error {
	channel, err := p.conn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Сериализуем сообщение в JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Публикуем с подтверждением доставки
	err = channel.PublishWithContext(
		ctx,
		"",        // exchange (пустой = default exchange)
		queueName, // routing key = имя очереди
		false,     // mandatory - не обязательно чтобы очередь существовала
		false,     // immediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent, // Persistent - сообщение переживёт перезапуск
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		p.logger.Error("Failed to publish message",
			zap.Error(err),
			zap.String("queue", queueName),
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Message published",
		zap.String("queue", queueName),
		zap.Int("size", len(body)),
	)

	return nil
}

// PublishWithDelay публикует сообщение с задержкой (для Sleep ноды)
// Требует плагин rabbitmq_delayed_message_exchange или использовать at library
func (p *Publisher) PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error {
	// TODO: Пока через at library будем делать
	// Когда интегрируем at - будем вызывать at.Schedule()
	
	p.logger.Warn("PublishWithDelay not implemented yet, use 'at' library",
		zap.String("queue", queueName),
		zap.Duration("delay", delay),
	)

	return fmt.Errorf("delayed publish not implemented, use 'at' library")
}
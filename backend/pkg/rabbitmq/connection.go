package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Connection управляет подключением к RabbitMQ
type Connection struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	url     string
	logger  *zap.Logger
	mu      sync.RWMutex
	closed  bool
}

// NewConnection создаёт новое подключение к RabbitMQ
func NewConnection(url string, logger *zap.Logger) (*Connection, error) {
	c := &Connection{
		url:    url,
		logger: logger,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	// Запускаем горутину для переподключения при разрыве
	go c.handleReconnect()

	return c, nil
}

// connect устанавливает соединение
func (c *Connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	// Подключаемся к RabbitMQ
	c.conn, err = amqp091.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Создаём канал
	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to create channel: %w", err)
	}

	c.logger.Info("Successfully connected to RabbitMQ")

	return nil
}

// handleReconnect переподключается при разрыве соединения
func (c *Connection) handleReconnect() {
	for {
		// Ждём сигнал о разрыве соединения
		err := <-c.conn.NotifyClose(make(chan *amqp091.Error))
		
		if c.closed {
			c.logger.Info("Connection closed gracefully")
			return
		}

		c.logger.Warn("RabbitMQ connection lost, reconnecting...", zap.Error(err))

		// Пытаемся переподключиться с экспоненциальной задержкой
		for i := 0; i < 10; i++ {
			time.Sleep(time.Duration(i+1) * time.Second)

			if err := c.connect(); err == nil {
				c.logger.Info("Successfully reconnected to RabbitMQ")
				break
			} else {
				c.logger.Warn("Failed to reconnect, retrying...", zap.Error(err))
			}
		}
	}
}

// GetChannel возвращает канал для работы
func (c *Connection) GetChannel() (*amqp091.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	return c.channel, nil
}

// Close закрывает соединение
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("Failed to close channel", zap.Error(err))
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Failed to close connection", zap.Error(err))
			return err
		}
	}

	c.logger.Info("RabbitMQ connection closed")
	return nil
}